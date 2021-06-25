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
	"fmt"
	"os"
	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	enableAutodiscoverEnvVar = "TNF_ENABLE_CONFIG_AUTODISCOVER"
	labelNamespace           = "test-network-function.com"
	labelTemplate            = "%s/%s"

	// anyLabelValue is the value that will allow any value for a label when building the label query.
	anyLabelValue = ""
)

// PerformAutoDiscovery checks the environment variable to see if autodiscovery should be performed
func PerformAutoDiscovery() (doAuto bool) {
	doAuto, _ = strconv.ParseBool(os.Getenv(enableAutodiscoverEnvVar))
	return doAuto
}

func buildLabelName(labelNS, labelName string) string {
	if labelNS == "" {
		return labelName
	}
	return fmt.Sprintf(labelTemplate, labelNS, labelName)
}

func buildAnnotationName(annotationName string) string {
	return buildLabelName(labelNamespace, annotationName)
}

func buildLabelQuery(labelNS, labelName, labelValue string) string {
	namespacedLabel := buildLabelName(labelNS, labelName)
	if labelValue != anyLabelValue {
		return fmt.Sprintf("%s=%s", namespacedLabel, labelValue)
	}
	return namespacedLabel
}

func makeGetCommand(resourceType, labelQuery string) *exec.Cmd {
	// TODO: shell expecter
	cmd := exec.Command("oc", "get", resourceType, "-A", "-o", "json", "-l", labelQuery)
	log.Debug("Issuing get command ", cmd.Args)

	return cmd
}
