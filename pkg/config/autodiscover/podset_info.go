// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

var (
	jsonUnmarshal     = json.Unmarshal
	execCommandOutput = func(command string) string {
		return utils.ExecuteCommandAndValidate(command, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
			log.Error("can't run command: ", command)
		})
	}
)

// Deploymen/statefulsettList holds the data from an `oc get deployment/statefulset -o json` command
type PodSetList struct {
	Items []PodSetResource `json:"items"`
}

// PodSetResource defines deployment/statefulset resources
type PodSetResource struct {
	Metadata struct {
		Name        string            `json:"name"`
		Namespace   string            `json:"namespace"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	} `json:"metadata"`

	Spec struct {
		Replicas int `json:"replicas"`
	}
}

// GetName returns podset's metadata section's name field.
func (podset *PodSetResource) GetName() string {
	return podset.Metadata.Name
}

// GetNamespace returns podset's metadata section's namespace field.
func (podset *PodSetResource) GetNamespace() string {
	return podset.Metadata.Namespace
}

// GetReplicas returns podset's spec section's replicas field.
func (podset *PodSetResource) GetReplicas() int {
	return podset.Spec.Replicas
}

// GetLabels returns a map with the podset's metadata section's labels.
func (podset *PodSetResource) GetLabels() map[string]string {
	return podset.Metadata.Labels
}
func (podset *PodSetResource) GetHpa() configsections.Hpa {
	template := fmt.Sprintf("go-template='{{ range .items }}{{ if eq .spec.scaleTargetRef.name %q }}{{.spec.minReplicas}},{{.spec.maxReplicas}},{{.metadata.name}}{{ end }}{{ end }}'", podset.GetName())
	ocCmd := fmt.Sprintf("oc get hpa -n %s -o %s", podset.GetNamespace(), template)
	out := execCommandOutput(ocCmd)
	if out != "" {
		out := strings.Split(out, ",")
		min, _ := strconv.Atoi(out[0])
		max, _ := strconv.Atoi(out[1])
		hpaNmae := out[2]
		return configsections.Hpa{
			MinReplicas: min,
			MaxReplicas: max,
			HpaName:     hpaNmae,
		}
	}
	return configsections.Hpa{}
}

// GetTargetPodSetsByNamespace will return all podsets(deployments/statefulset )that have pods with a given label.
func GetTargetPodSetsByNamespace(namespace string, targetLabel configsections.Label, resourceTypePodSet string) (*PodSetList, error) {
	labelQuery := fmt.Sprintf("%q==%q", buildLabelName(targetLabel.Prefix, targetLabel.Name), targetLabel.Value)
	jqArgs := fmt.Sprintf("'[.items[] | select(.spec.template.metadata.labels.%s)]'", labelQuery)
	ocCmd := fmt.Sprintf("oc get %s -n %s -o json | jq %s", resourceTypePodSet, namespace, jqArgs)

	out := execCommandOutput(ocCmd)

	var podsetList PodSetList
	err := jsonUnmarshal([]byte(out), &podsetList.Items)
	if err != nil {
		return nil, err
	}

	return &podsetList, nil
}

// GetTargetDeploymentsByLabel will return all deployments/statefulsets that have pods with a given label.
func GetTargetPodSetsByLabel(targetLabel configsections.Label, resourceTypePodSet string) (*PodSetList, error) {
	labelQuery := fmt.Sprintf("%q==%q", buildLabelName(targetLabel.Prefix, targetLabel.Name), targetLabel.Value)
	jqArgs := fmt.Sprintf("'[.items[] | select(.spec.template.metadata.labels.%s)]'", labelQuery)
	ocCmd := fmt.Sprintf("oc get %s -A -o json | jq %s", resourceTypePodSet, jqArgs)

	out := execCommandOutput(ocCmd)

	var podsetList PodSetList
	err := jsonUnmarshal([]byte(out), &podsetList.Items)
	if err != nil {
		return nil, err
	}

	return &podsetList, nil
}
