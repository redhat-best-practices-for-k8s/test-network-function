// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"time"

	"github.com/onsi/ginkgo"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/internal/api"
	configpkg "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	// timeout for eventually call
	apiRequestTimeout = 30 * time.Second
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
		ginkgo.AfterEach(env.CloseLocalShellContext)

		testContainerCertificationStatus()
		testOperatorCertificationStatus()
	}
})

// getContainerCertificationRequestFunction returns function that will try to get the certification status (CCP) for a container.
func getContainerCertificationRequestFunction(repository, containerName string) func() bool {
	return func() bool {
		return certAPIClient.IsContainerCertified(repository, containerName)
	}
}

// getOperatorCertificationRequestFunction returns function that will try to get the certification status (OCP) for an operator.
func getOperatorCertificationRequestFunction(organization, operatorName string) func() bool {
	return func() bool {
		return certAPIClient.IsOperatorCertified(organization, operatorName)
	}
}

// waitForCertificationRequestToSuccess calls to certificationRequestFunc until it returns true.
func waitForCertificationRequestToSuccess(certificationRequestFunc func() bool, timeout time.Duration) bool {
	const pollingPeriod = 1 * time.Second
	var elapsed time.Duration
	isCertified := false

	for elapsed < timeout {
		isCertified = certificationRequestFunc()

		if isCertified {
			break
		}
		time.Sleep(pollingPeriod)
		elapsed += pollingPeriod
	}
	return isCertified
}

func testContainerCertificationStatus() {
	// Query API for certification status of listed containers
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestContainerIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		env := configpkg.GetTestEnvironment()
		containersToQuery := make(map[configsections.CertifiedContainerRequestInfo]bool)
		for _, c := range env.Config.CertifiedContainerInfo {
			containersToQuery[c] = true
		}
		if env.Config.CheckDiscoveredContainerCertificationStatus {
			for _, cut := range env.ContainersUnderTest {
				containersToQuery[configsections.CertifiedContainerRequestInfo{Repository: cut.ImageSource.Repository, Name: cut.ImageSource.Name}] = true
			}
		}
		if len(containersToQuery) == 0 {
			ginkgo.Skip("No containers to check configured in tnf_config.yml")
		}
		ginkgo.By(fmt.Sprintf("Getting certification status. Number of containers to check: %d", len(containersToQuery)))
		if len(containersToQuery) > 0 {
			certAPIClient = api.NewHTTPClient()
			failedContainers := []configsections.CertifiedContainerRequestInfo{}
			allContainersToQueryEmpty := true
			for c := range containersToQuery {
				if c.Name == "" || c.Repository == "" {
					tnf.ClaimFilePrintf("Container name = \"%s\" or repository = \"%s\" is missing, skipping this container to query", c.Name, c.Repository)
					continue
				}
				allContainersToQueryEmpty = false
				ginkgo.By(fmt.Sprintf("Container %s/%s should eventually be verified as certified", c.Repository, c.Name))
				isCertified := waitForCertificationRequestToSuccess(getContainerCertificationRequestFunction(c.Repository, c.Name), apiRequestTimeout)
				if !isCertified {
					tnf.ClaimFilePrintf("Container %s (repository %s) is not found in the certified container catalog.", c.Name, c.Repository)
					failedContainers = append(failedContainers, c)
				} else {
					log.Info(fmt.Sprintf("Container %s (repository %s) is certified.", c.Name, c.Repository))
				}
			}
			if allContainersToQueryEmpty {
				ginkgo.Skip("No containers to check because either container name or repository is empty for all containers in tnf_config.yml")
			}

			if n := len(failedContainers); n > 0 {
				log.Warnf("Containers that are not certified: %+v", failedContainers)
				ginkgo.Fail(fmt.Sprintf("%d container images are not certified.", n))
			}
		}
	})
}

func testOperatorCertificationStatus() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		operatorsToQuery := configpkg.GetTestEnvironment().Config.CertifiedOperatorInfo

		if len(operatorsToQuery) == 0 {
			ginkgo.Skip("No operators to check configured in tnf_config.yml")
		}

		ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators to check: %d", len(operatorsToQuery)))
		if len(operatorsToQuery) > 0 {
			certAPIClient = api.NewHTTPClient()
			failedOperators := []configsections.CertifiedOperatorRequestInfo{}
			allOperatorsToQueryEmpty := true
			for _, operator := range operatorsToQuery {
				if operator.Name == "" || operator.Organization == "" {
					tnf.ClaimFilePrintf("Operator name = \"%s\" or organization = \"%s\" is missing, skipping this operator to query", operator.Name, operator.Organization)
					continue
				}
				allOperatorsToQueryEmpty = false
				ginkgo.By(fmt.Sprintf("Should eventually be verified as certified (operator %s/%s)", operator.Organization, operator.Name))
				isCertified := waitForCertificationRequestToSuccess(getOperatorCertificationRequestFunction(operator.Organization, operator.Name), apiRequestTimeout)
				if !isCertified {
					tnf.ClaimFilePrintf("Operator %s (organization %s) failed to be certified.", operator.Name, operator.Organization)
					failedOperators = append(failedOperators, operator)
				} else {
					log.Info(fmt.Sprintf("Operator %s (organization %s) certified OK.", operator.Name, operator.Organization))
				}
			}
			if allOperatorsToQueryEmpty {
				ginkgo.Skip("No operators to check because either operator name or organization is empty for all operators in tnf_config.yml")
			}

			if n := len(failedOperators); n > 0 {
				log.Warnf("Operators that failed to be certified: %+v", failedOperators)
				ginkgo.Fail(fmt.Sprintf("%d operators failed to be certified.", n))
			}
		}
	})
}
