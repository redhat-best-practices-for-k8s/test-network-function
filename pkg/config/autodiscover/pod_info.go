// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package autodiscover

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

const (
	cnfDefaultNetworkInterfaceKey = "defaultnetworkinterface"
	cniNetworksStatusKey          = "k8s.v1.cni.cncf.io/networks-status"
	resourceTypePods              = "pods"
	podPhaseRunning               = "Running"
)

var (
	namespacedDefaultNetworkInterfaceKey = buildAnnotationName(cnfDefaultNetworkInterfaceKey)
)

// PodList holds the data from an `oc get pods -o json` command
type PodList struct {
	Items []*PodResource `json:"items"`
}

// PodResource is a single Pod from an `oc get pods -o json` command
type PodResource struct {
	Metadata struct {
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		DeletionTimestamp string            `json:"deletionTimestamp"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
	} `json:"metadata"`
	Spec struct {
		ServiceAccount string `json:"serviceaccountname"`
		Containers     []struct {
			Name string `json:"name"`
		} `json:"containers"`
		NodeName string `json:"nodeName"`
	} `json:"spec"`
	Status struct {
		// PodIPs this is currently unused, but part of the oc get output
		// The listof IPs is contained in the Metadata->Annotations section
		// of this structure. This is a list of ips with the following format:
		// [0]:map[string]string ["ip": "10.130.0.65", ]
		PodIPs            []map[string]string `json:"podIPs"`
		Phase             string              `json:"phase"`
		ContainerStatuses []struct {
			Name        string `json:"name"`
			ContainerID string `json:"containerID"`
		} `json:"containerStatuses"`
	} `json:"status"`
}

type cniNetworkInterface struct {
	Name      string                 `json:"name"`
	Interface string                 `json:"interface"`
	IPs       []string               `json:"ips"`
	Default   bool                   `json:"default"`
	DNS       map[string]interface{} `json:"dns"`
}

func (pr *PodResource) hasAnnotation(annotationKey string) (present bool) {
	_, present = pr.Metadata.Annotations[annotationKey]
	return
}

// GetAnnotationValue will get the value stored in the given annotation and
// Unmarshal it into the given var `v`.
func (pr *PodResource) GetAnnotationValue(annotationKey string, v interface{}) (err error) {
	if !pr.hasAnnotation(annotationKey) {
		return fmt.Errorf("failed to find annotation '%s' on pod '%s/%s'", annotationKey, pr.Metadata.Namespace, pr.Metadata.Name)
	}
	val := pr.Metadata.Annotations[annotationKey]
	err = jsonUnmarshal([]byte(val), v)
	if err != nil {
		return pr.annotationUnmarshalError(annotationKey, err)
	}
	return
}

// getDefaultNetworkDeviceFromAnnotations examins the pod annotations to try and determine the primary network device.
// First, if the cnf-certification-specific annotation "test-network-function.com/defaultnetworkinterface" is present
// then the value of that will be decoded and returned. It must be a single JSON-encoded string.
// Next, if the "k8s.v1.cni.cncf.io/networks-status" annotation is present then the first entry where `default == true`
// will be used. Note that this annotation may not be present outside OpenShift.  Returns (interface, error).
func (pr *PodResource) getDefaultNetworkDeviceFromAnnotations() (string, error) {
	// Note: The `GetAnnotationValue` method does not distinguish between bad encoding and a missing annotation, which is needed here.
	var iface string
	if val, present := pr.Metadata.Annotations[namespacedDefaultNetworkInterfaceKey]; present {
		err := jsonUnmarshal([]byte(val), &iface)
		return iface, err
	}
	if val, present := pr.Metadata.Annotations[cniNetworksStatusKey]; present {
		var cniInfo []cniNetworkInterface
		err := jsonUnmarshal([]byte(val), &cniInfo)
		if err != nil {
			return "", pr.annotationUnmarshalError(cniNetworksStatusKey, err)
		}
		for _, cniInterface := range cniInfo {
			if cniInterface.Default {
				return cniInterface.Interface, nil
			}
		}
	}
	return "", fmt.Errorf("unable to determine a default network interface for %s/%s", pr.Metadata.Namespace, pr.Metadata.Name)
}

// getPodIPsPerNet gets the IPs of a pod.
// CNI annotation "k8s.v1.cni.cncf.io/networks-status".
// Returns (ips, error).
func (pr *PodResource) getPodIPsPerNet() (map[string][]string, error) {
	// This is a map indexed with the network name (network attachment) and
	// listing all the IPs created in this subnet and belonging to the pod namespace
	// The list of ips pr net is parsed from the content of the "k8s.v1.cni.cncf.io/networks-status" annotation.
	// see file pkg/config/autodiscover/testdata/testtarget.json for an example of such annotation
	ips := make(map[string][]string)

	if val, present := pr.Metadata.Annotations[cniNetworksStatusKey]; present {
		var cniInfo []cniNetworkInterface
		err := jsonUnmarshal([]byte(val), &cniInfo)
		if err != nil {
			return nil, pr.annotationUnmarshalError(cniNetworksStatusKey, err)
		}
		// If this is the default interface, skip it as it is tested separately
		// Otherwise add all non default interfaces
		for _, cniInterface := range cniInfo {
			if !cniInterface.Default {
				ips[cniInterface.Name] = cniInterface.IPs
			}
		}
		return ips, nil
	}
	log.Warn("Could not establish pod IPs from annotations, please manually set the 'test-network-function.com/multusips' annotation for complete test coverage")

	return ips, nil
}

func (pr *PodResource) annotationUnmarshalError(annotationKey string, err error) error {
	return fmt.Errorf("error (%s) attempting to unmarshal value of annotation '%s' on pod '%s/%s'",
		err, annotationKey, pr.Metadata.Namespace, pr.Metadata.Name)
}

// GetPodsByLabelByNamespace will return all pods with a given label value in provided namespace.
// If `labelValue` is an empty string, all pods with that
// label will be returned, regardless of the labels value.
func GetPodsByLabelByNamespace(label configsections.Label, namespace string) (*PodList, error) {
	out := executeOcGetCommand(resourceTypePods, buildLabelQuery(label), namespace)

	log.Debug("JSON output for all pods labeled with: ", label)
	log.Debug("Command: ", out)

	var podList PodList
	err := jsonUnmarshal([]byte(out), &podList)
	if err != nil {
		return nil, err
	}

	// Filter out terminating pods and pending/unscheduled pods
	var pods []*PodResource
	for _, pod := range podList.Items {
		if pod.Metadata.DeletionTimestamp == "" || pod.Status.Phase != podPhaseRunning {
			pods = append(pods, pod)
		}
	}
	podList.Items = pods
	return &podList, nil
}

// GetPodsByLabelByNamespace will return all pods with a given label value.
// If `labelValue` is an empty string, all pods with that
// label will be returned, regardless of the labels value.
func GetPodsByLabel(label configsections.Label) (*PodList, error) {
	out := executeOcGetAllCommand(resourceTypePods, buildLabelQuery(label))

	log.Debug("JSON output for all pods labeled with: ", label)
	log.Debug("Command: ", out)

	var podList PodList
	err := jsonUnmarshal([]byte(out), &podList)
	if err != nil {
		return nil, err
	}

	// Filter out terminating pods and pending/unscheduled pods
	var pods []*PodResource
	for _, pod := range podList.Items {
		if pod.Metadata.DeletionTimestamp == "" || pod.Status.Phase != podPhaseRunning {
			pods = append(pods, pod)
		}
	}
	podList.Items = pods
	return &podList, nil
}
