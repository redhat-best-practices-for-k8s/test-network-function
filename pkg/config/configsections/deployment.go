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

package configsections

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

var (
	ocCommandTimeOut            = time.Second * 10
	expectersVerboseModeEnabled = false
	execCommandOutput           = func(command string) string {
		return utils.ExecuteCommand(command, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
			log.Error("can't run command: ", command)
		})
	}
)

// Deployment defines a deployment in the cluster.
type Deployment struct {
	Name      string
	Namespace string
	Replicas  int
}

func (deployment Deployment) IsHpa() Hpa {
	template := fmt.Sprintf("go-template='{{ range .items }}{{ if eq .spec.scaleTargetRef.name \"%s\" }}{{.spec.minReplicas}},{{.spec.maxReplicas}},{{.metadata.name}}{{ end }}{{ end }}'", deployment.Name)
	ocCmd := fmt.Sprintf("oc get hpa -n %s -o %s", deployment.Namespace, template)
	out := execCommandOutput(ocCmd)
	if out != "" {
		out := strings.Split(out, ",")
		min, _ := strconv.Atoi(out[0])
		max, _ := strconv.Atoi(out[1])
		hpaNmae := out[2]
		return Hpa{
			MinReplicas: min,
			MaxReplicas: max,
			HpaName:     hpaNmae,
		}
	}
	return Hpa{
		MinReplicas: 0,
		MaxReplicas: 0,
		HpaName:     "",
	}
}
