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

	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

const (
	resourceTypeDeployment = "deployment"
)

// DeploymentList holds the data from an `oc get deployments -o json` command
type DeploymentList struct {
	Items []DeploymentResource `json:"items"`
}

// DeploymentResource defines deployment resources
type DeploymentResource struct {
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

// GetName returns deployment's metadata section's name field.
func (deployment *DeploymentResource) GetName() string {
	return deployment.Metadata.Name
}

// GetNamespace returns deployment's metadata section's namespace field.
func (deployment *DeploymentResource) GetNamespace() string {
	return deployment.Metadata.Namespace
}

// GetReplicas returns deployment's spec section's replicas field.
func (deployment *DeploymentResource) GetReplicas() int {
	return deployment.Spec.Replicas
}

// GetLabels returns a map with the deployment's metadata section's labels.
func (deployment *DeploymentResource) GetLabels() map[string]string {
	return deployment.Metadata.Labels
}

// GetTargetDeploymentsByNamespace will return all pods with a given label value. If `labelValue` is an empty string, all pods with that
// label will be returned, regardless of the labels value.
func GetTargetDeploymentsByNamespace(namespace string) (*DeploymentList, error) {
	// Filter out the deployments that match the exclusionLabels, currently the orchestrator and the fs_diff ones.
	exclusionLabels := []configsections.Label{
		{Namespace: tnfNamespace, Name: genericLabelName, Value: orchestratorValue},
		{Namespace: tnfNamespace, Name: genericLabelName, Value: fsDiffMasterValue}}

	cmd := makeGetCommandByNamespace(namespace, resourceTypeDeployment, buildExclusionLabelsQuery(exclusionLabels))

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var deploymentList DeploymentList
	err = json.Unmarshal(out, &deploymentList)
	if err != nil {
		return nil, err
	}

	return &deploymentList, nil
}
