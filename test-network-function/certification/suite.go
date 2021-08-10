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
	ginkgoconfig "github.com/onsi/ginkgo/config"
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
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.AffiliatedCertTestKey) {

		// Query API for certification status of listed containers
		ginkgo.When("getting certification status", func() {
			conf := configpkg.GetConfigInstance()
			cnfsToQuery := conf.CertifiedContainerInfo
			if len(cnfsToQuery) > 0 {
				certAPIClient = api.NewHTTPClient()
				for _, cnfRequestInfo := range cnfsToQuery {
					cnf := cnfRequestInfo
					// Care: this test takes some time to run, failures at later points while before this has finished may be reported as a failure here. Read the failure reason carefully.
					ginkgo.It(fmt.Sprintf("container %s/%s should eventually be verified as certified", cnf.Repository, cnf.Name), func() {
						defer results.RecordResult(identifiers.TestContainerIsCertifiedIdentifier)
						cnf := cnf // pin
						gomega.Eventually(func() bool {
							isCertified := certAPIClient.IsContainerCertified(cnf.Repository, cnf.Name)
							return isCertified
						}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
					})
				}
			}
		})

		operatorsToQuery := configpkg.GetConfigInstance().CertifiedOperatorInfo
		if len(operatorsToQuery) > 0 {
			certAPIClient := api.NewHTTPClient()
			for _, certified := range operatorsToQuery {
				// Care: this test takes some time to run, failures at later points while before this has finished may be reported as a failure here. Read the failure reason carefully.
				ginkgo.It(fmt.Sprintf("should eventually be verified as certified (operator %s/%s)", certified.Organization, certified.Name), func() {
					defer results.RecordResult(identifiers.TestOperatorIsCertifiedIdentifier)
					certified := certified // pin
					gomega.Eventually(func() bool {
						isCertified := certAPIClient.IsOperatorCertified(certified.Organization, certified.Name)
						return isCertified
					}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
				})
			}
		}

	}
})
