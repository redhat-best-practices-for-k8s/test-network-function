// Copyright (C) 2021 Red Hat, Inc.
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
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

const (
	genericLabelName  = "generic"
	orchestratorValue = "orchestrator"
	defaultNamespace  = "default"
	debugLabelName    = "app"
	debugLabelValue   = "debug"
)

// FindTestPartner completes a `configsections.TestPartner` from the current state of the cluster,
// using labels and annotations to populate the data, if it's not fully configured
func FindTestPartner(tp *configsections.TestPartner, namespace string) {
	if tp.TestOrchestratorID.ContainerName == "" {
		orchestrator, err := getContainerByLabel(configsections.Label{Prefix: tnfLabelPrefix, Name: genericLabelName, Value: orchestratorValue}, namespace)
		if err != nil {
			log.Errorf("failed to identify a single test orchestrator container: %s", err)
			return
		}
		tp.ContainerConfigList = append(tp.ContainerConfigList, orchestrator)
		tp.TestOrchestratorID = orchestrator.ContainerIdentifier
	}
	label := configsections.Label{Name: debugLabelName, Value: debugLabelValue}
	pods, err := GetPodsByLabel(label, defaultNamespace)
	if err != nil {
		log.Panic("can't find debug pods")
	}
	if len(pods.Items) == 0 {
		log.Panic("can't find debug pods, make sure daemonset debug is deployed properly")
	}
	for _, pod := range pods.Items {
		tp.ContainersDebugList = append(tp.ContainersDebugList, buildContainersFromPodResource(pod)[0])
	}
}
