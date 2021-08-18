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
	fsDiffMasterValue = "fs_diff_master"
)

// FindTestPartner completes a `configsections.TestPartner` from the current state of the cluster,
// using labels and annotations to populate the data, if it's not fully configured
func FindTestPartner(tp *configsections.TestPartner) {
	if tp.TestOrchestratorID.ContainerName == "" {
		orchestrator, err := getContainerByLabel(configsections.Label{Namespace: tnfNamespace, Name: genericLabelName, Value: orchestratorValue})
		if err != nil {
			log.Fatalf("failed to identify a single test orchestrator container: %s", err)
		}
		tp.ContainerConfigList = append(tp.ContainerConfigList, orchestrator)
		tp.TestOrchestratorID = orchestrator.ContainerIdentifier
	}

	if tp.FsDiffMasterContainerID.ContainerName == "" {
		fsDiffMasterContainer, err := getContainerByLabel(configsections.Label{Namespace: tnfNamespace, Name: genericLabelName, Value: fsDiffMasterValue})
		if err == nil {
			tp.ContainerConfigList = append(tp.ContainerConfigList, fsDiffMasterContainer)
			tp.FsDiffMasterContainerID = fsDiffMasterContainer.ContainerIdentifier
		} else {
			log.Warnf("an error (%s) occurred when getting the FS Diff Master Container. Attempting to continue", err)
		}
	}
}
