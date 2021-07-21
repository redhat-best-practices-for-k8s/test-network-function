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

package generic

import (
	"fmt"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/serviceaccount"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

var _ = ginkgo.Describe(accessControlTestKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, accessControlTestKey) {
		config := GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		containersUnderTest := createContainersUnderTest(config)
		// partnerContainers := createPartnerContainers(config)
		// testOrchestrator := partnerContainers[config.TestOrchestrator]

		// log.Info(testOrchestrator)
		// log.Info(containersUnderTest)

		for _, containerUnderTest := range containersUnderTest {
			testNamespace(containerUnderTest.oc)
		}

		for _, containerUnderTest := range containersUnderTest {
			testRoles(containerUnderTest.oc.GetPodName(), containerUnderTest.oc.GetPodNamespace())
		}

	}
})

func testNamespace(oc *interactive.Oc) {
	pod := oc.GetPodName()
	container := oc.GetPodContainerName()
	ginkgo.When(fmt.Sprintf("Reading namespace of %s/%s", pod, container), func() {
		ginkgo.It("Should not be 'default' and should not begin with 'openshift-'", func() {
			defer results.RecordResult(identifiers.TestNamespaceBestPracticesIdentifier)
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.Equal("default")))
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.HavePrefix("openshift-")))
		})
	})
}

func testRoles(podName, podNamespace string) {
	var serviceAccountName string
	ginkgo.When(fmt.Sprintf("Testing roles and privileges of %s/%s", podNamespace, podName), func() {
		testServiceAccount(podName, podNamespace, &serviceAccountName)
		testRoleBindings(podNamespace, &serviceAccountName)
		testClusterRoleBindings(podNamespace, &serviceAccountName)
	})
}

func testServiceAccount(podName, podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should have a valid ServiceAccount name", func() {
		defer results.RecordResult(identifiers.TestPodServiceAccountBestPracticesIdentifier)
		context := getContext()
		tester := serviceaccount.NewServiceAccount(defaultTimeout, podName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
		*serviceAccountName = tester.GetServiceAccountName()
		gomega.Expect(*serviceAccountName).ToNot(gomega.BeEmpty())
	})
}

func testRoleBindings(podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have RoleBinding in other namespaces", func() {
		defer results.RecordResult(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		context := getContext()
		rbTester := rolebinding.NewRoleBinding(defaultTimeout, *serviceAccountName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), rbTester, []reel.Handler{rbTester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		if rbTester.Result() == tnf.FAILURE {
			log.Info("RoleBindings: ", rbTester.GetRoleBindings())
		}
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
	})
}

func testClusterRoleBindings(podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have ClusterRoleBindings", func() {
		defer results.RecordResult(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		context := getContext()
		crbTester := clusterrolebinding.NewClusterRoleBinding(defaultTimeout, *serviceAccountName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), crbTester, []reel.Handler{crbTester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		if crbTester.Result() == tnf.FAILURE {
			log.Info("ClusterRoleBindings: ", crbTester.GetClusterRoleBindings())
		}
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	})
}
