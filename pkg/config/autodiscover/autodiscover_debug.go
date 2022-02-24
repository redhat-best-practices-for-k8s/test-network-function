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
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	ds "github.com/test-network-function/test-network-function/pkg/tnf/handlers/daemonset"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

const (
	defaultNamespace   = "default"
	debugDaemonSet     = "debug"
	debugLabelName     = "test-network-function.com/app"
	debugLabelValue    = "debug"
	nodeLabelName      = "test-network-function.com/node"
	nodeLabelValue     = "target"
	addlabelCommand    = "oc label node %s %s=%s --overwrite=true"
	deletelabelCommand = "oc label node %s %s- --overwrite=true"
)

// FindDebugPods completes a `configsections.TestPartner.ContainersDebugList` from the current state of the cluster,
// using labels and annotations to populate the data, if it's not fully configured
func FindDebugPods(tp *configsections.TestPartner) {
	label := configsections.Label{Name: debugLabelName, Value: debugLabelValue}
	pods, err := GetPodsByLabelByNamespace(label, defaultNamespace)
	if err != nil {
		log.Panic("can't find debug pods. Error: ", err)
	}
	if len(pods.Items) == 0 {
		log.Panic("can't find debug pods, make sure daemonset debug is deployed properly")
	}
	for _, pod := range pods.Items {
		tp.ContainersDebugList = append(tp.ContainersDebugList, buildContainers(pod)[0])
	}
}

// AddDebugLabel add debug label to node
func AddDebugLabel(nodeName string) {
	log.Info("add label ", nodeLabelName, "=", nodeLabelValue, " to node ", nodeName)
	ocCommand := fmt.Sprintf(addlabelCommand, nodeName, nodeLabelName, nodeLabelValue)
	_ = utils.ExecuteCommandAndValidate(ocCommand, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
		log.Error("error in adding label to node ", nodeName)
	})
}

// AddDebugLabel remove debug label from node
func DeleteDebugLabel(nodeName string) {
	log.Info("delete label ", nodeLabelName, "=", nodeLabelValue, "to node ", nodeName)
	ocCommand := fmt.Sprintf(deletelabelCommand, nodeName, nodeLabelName)
	_ = utils.ExecuteCommandAndValidate(ocCommand, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
		log.Error("error in removing label from node ", nodeName)
	})
}

// CheckDebugDaemonset checks if the debug pods are deployed properly
// the function will try DefaultTimeout/time.Second times
func CheckDebugDaemonset(expectedDebugPods int) {
	gomega.Eventually(func() bool {
		log.Debug("check debug daemonset status")
		return checkDebugPodsReadiness(expectedDebugPods)
	}, 60*time.Second, 2*time.Second).Should(gomega.Equal(true)) //nolint: gomnd
}

// checkDebugPodsReadiness helper function that returns true if the daemonset debug is deployed properly
func checkDebugPodsReadiness(expectedDebugPods int) bool {
	context := interactive.GetContext(expectersVerboseModeEnabled)
	tester := ds.NewDaemonSet(DefaultTimeout, debugDaemonSet, defaultNamespace)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	if err != nil {
		log.Error("can't run test to detect daemonset status")
		return false
	}
	_, err = test.Run()
	if err != nil {
		return false
	}
	dsStatus := tester.GetStatus()
	if expectedDebugPods == dsStatus.Desired &&
		dsStatus.Desired == dsStatus.Current &&
		dsStatus.Current == dsStatus.Available &&
		dsStatus.Available == dsStatus.Ready &&
		dsStatus.Misscheduled == 0 {
		log.Info("daemonset is ready")
		return true
	}
	log.Warn("daemonset is not ready")
	return false
}
