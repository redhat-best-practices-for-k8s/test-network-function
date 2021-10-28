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

package certification

import (
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/internal/api"
	configpkg "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	// timeout for eventually call
	eventuallyTimeoutSeconds = 30
	// interval of time
	interval = 1
)

var certAPIClient api.CertAPIClient

var _ = ginkgo.Describe(common.AffiliatedCertTestKey, func() {
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.AffiliatedCertTestKey) {
		env := configpkg.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
		})

		ginkgo.ReportAfterEach(results.RecordResult)

		testContainerCertificationStatus()
		testOperatorCertificationStatus()
	}
})

func testContainerCertificationStatus() {
	// Query API for certification status of listed containers
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestContainerIsCertifiedIdentifier)
	ginkgo.It(testID, func() {
		env := configpkg.GetTestEnvironment()
		cnfsToQuery := env.Config.CertifiedContainerInfo

		ginkgo.By(fmt.Sprintf("Getting certification status. Number of containers to check: %d", len(cnfsToQuery)))

		if len(cnfsToQuery) > 0 {
			certAPIClient = api.NewHTTPClient()
			for _, cnf := range cnfsToQuery {
				cnf := cnf // pin
				// Care: this test takes some time to run, failures at later points while before this has finished may be reported as a failure here. Read the failure reason carefully.
				ginkgo.By(fmt.Sprintf("container %s/%s should eventually be verified as certified", cnf.Repository, cnf.Name))
				gomega.Eventually(func() bool {
					isCertified := certAPIClient.IsContainerCertified(cnf.Repository, cnf.Name)
					return isCertified
				}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
			}
		}
	})
}

func testOperatorCertificationStatus() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, func() {
		operatorsToQuery := configpkg.GetTestEnvironment().Config.CertifiedOperatorInfo
		ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators to check: %d", len(operatorsToQuery)))
		if len(operatorsToQuery) > 0 {
			certAPIClient := api.NewHTTPClient()
			for _, certified := range operatorsToQuery {
				ginkgo.By(fmt.Sprintf("should eventually be verified as certified (operator %s/%s)", certified.Organization, certified.Name))
				// Care: this test takes some time to run, failures at later points while before this has finished may be reported as a failure here. Read the failure reason carefully.
				certified := certified // pin
				gomega.Eventually(func() bool {
					isCertified := certAPIClient.IsOperatorCertified(certified.Organization, certified.Name)
					return isCertified
				}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
			}
		}
	})
}
