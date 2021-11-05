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
	cnfIPsKey                     = "multusips"
	cniNetworksStatusKey          = "k8s.v1.cni.cncf.io/networks-status"
	resourceTypePods              = "pods"
	podPhaseRunning               = "Running"
)

var (
	namespacedDefaultNetworkInterfaceKey = buildAnnotationName(cnfDefaultNetworkInterfaceKey)
	namespacedIPsKey                     = buildAnnotationName(cnfIPsKey)
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
		PodIPs []map[string]string `json:"podIPs"`
		Phase  string              `json:"phase"`
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
// will be used. Note that this annotation may not be present outside OpenShift.
func (pr *PodResource) getDefaultNetworkDeviceFromAnnotations() (iface string, err error) {
	// Note: The `GetAnnotationValue` method does not distinguish between bad encoding and a missing annotation, which is needed here.
	if val, present := pr.Metadata.Annotations[namespacedDefaultNetworkInterfaceKey]; present {
		err = jsonUnmarshal([]byte(val), &iface)
		return
	}
	if val, present := pr.Metadata.Annotations[cniNetworksStatusKey]; present {
		var cniInfo []cniNetworkInterface
		err = jsonUnmarshal([]byte(val), &cniInfo)
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

// getPodIPs gets the IPs of a pod.
// In precedence, it uses the cnf-certification specific annotation if it is present. If set,
// "test-network-function.com/multusips" must be a json-encoded list of string IPs.
// The fallback option is the
// CNI annotation "k8s.v1.cni.cncf.io/networks-status". If neither are available, then `pod.status.ips` is used, though
// this may not contain all IPs in all cases and will not be _only_ multus IPs.
func (pr *PodResource) getPodIPs() (ips []string, err error) {
	// Note: The `GetAnnotationValue` method does not distinguish between bad encoding and a missing annotation, which is needed here.
	if val, present := pr.Metadata.Annotations[namespacedIPsKey]; present {
		err = jsonUnmarshal([]byte(val), &ips)
		return
	}
	if val, present := pr.Metadata.Annotations[cniNetworksStatusKey]; present {
		var cniInfo []cniNetworkInterface
		err = jsonUnmarshal([]byte(val), &cniInfo)
		if err != nil {
			return nil, pr.annotationUnmarshalError(cniNetworksStatusKey, err)
		}
		// If this is the default interface, skip it as it is tested separately
		// Otherwise add all non default interfaces
		for _, cniInterface := range cniInfo {
			if !cniInterface.Default {
				ips = append(ips, cniInterface.IPs...)
			}
		}
		return
	}
	log.Warn("Could not establish pod IPs from annotations, please manually set the 'test-network-function.com/multusips' annotation for complete test coverage")

	return
}

func (pr *PodResource) annotationUnmarshalError(annotationKey string, err error) error {
	return fmt.Errorf("error (%s) attempting to unmarshal value of annotation '%s' on pod '%s/%s'",
		err, annotationKey, pr.Metadata.Namespace, pr.Metadata.Name)
}

// GetPodsByLabel will return all pods with a given label value. If `labelValue` is an empty string, all pods with that
// label will be returned, regardless of the labels value.
func GetPodsByLabel(label configsections.Label, namespace string) (*PodList, error) {
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
