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
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/internal/api"
	configpkg "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/utils"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	// timeout for eventually call
	apiRequestTimeout           = 30 * time.Second
	expectersVerboseModeEnabled = false
	CertifiedOperator           = "certified-operators"
)

var (
	ocpVersionCommand = "oc version -o json | jq '.openshiftVersion'"

	execCommandOutput = func(command string) string {
		return utils.ExecuteCommandAndValidate(command, apiRequestTimeout, interactive.GetContext(expectersVerboseModeEnabled), func() {
			log.Error("can't run command: ", command)
		})
	}

	certAPIClient api.CertAPIClient
	ocpversion    = GetOcpVersion()
)

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
		testAllOperatorCertified(env)
	}
})

// getContainerCertificationRequestFunction returns function that will try to get the certification status (CCP) for a container.
func getContainerCertificationRequestFunction(id configsections.ContainerImageIdentifier) func() (bool, error) {
	return func() (bool, error) {
		return certAPIClient.IsContainerCertified(id)
	}
}

// getOperatorCertificationRequestFunction returns function that will try to get the certification status (OCP) for an operator.
func getOperatorCertificationRequestFunction(organization, operatorName string) func() (bool, error) {
	return func() (bool, error) {
		return certAPIClient.IsOperatorCertified(organization, operatorName, ocpversion)
	}
}

// waitForCertificationRequestToSuccess calls to certificationRequestFunc until it returns true.
func waitForCertificationRequestToSuccess(certificationRequestFunc func() (bool, error), timeout time.Duration) bool {
	const pollingPeriod = 1 * time.Second
	var elapsed time.Duration
	var err error
	isCertified := false

	for elapsed < timeout {
		isCertified, err = certificationRequestFunc()

		if err == nil {
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
		containersToQuery := make(map[configsections.ContainerImageIdentifier]bool)
		for _, c := range env.Config.CertifiedContainerInfo {
			containersToQuery[c] = true
		}
		if env.Config.CheckDiscoveredContainerCertificationStatus {
			for _, cut := range env.ContainersUnderTest {
				containersToQuery[cut.ImageSource.ContainerImageIdentifier] = true
			}
		}
		if len(containersToQuery) == 0 {
			ginkgo.Skip("No containers to check configured in tnf_config.yml")
		}
		ginkgo.By(fmt.Sprintf("Getting certification status. Number of containers to check: %d", len(containersToQuery)))
		if len(containersToQuery) > 0 {
			certAPIClient = api.NewHTTPClient()
			failedContainers := []configsections.ContainerImageIdentifier{}
			allContainersToQueryEmpty := true
			for c := range containersToQuery {
				if c.Name == "" || c.Repository == "" {
					tnf.ClaimFilePrintf("Container name = \"%s\" or repository = \"%s\" is missing, skipping this container to query", c.Name, c.Repository)
					continue
				}
				allContainersToQueryEmpty = false
				ginkgo.By(fmt.Sprintf("Container %s/%s should eventually be verified as certified", c.Repository, c.Name))
				isCertified := waitForCertificationRequestToSuccess(getContainerCertificationRequestFunction(c), apiRequestTimeout)
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

func testAllOperatorCertified(env *configpkg.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		operatorsToQuery := env.OperatorsUnderTest

		if len(operatorsToQuery) == 0 {
			ginkgo.Skip("No operators to check configured ")
		}

		ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators drivers to check: %d", len(operatorsToQuery)))

		testFailed := false
		for _, op := range operatorsToQuery {
			pack := op.Name
			org := op.Org
			if org == CertifiedOperator {
				isCertified := waitForCertificationRequestToSuccess(getOperatorCertificationRequestFunction(org, pack), apiRequestTimeout)
				if !isCertified {
					testFailed = true
					log.Info(fmt.Sprintf("Operator %s (organization %s) not certified because of the wrong version of the operator or the ocp version %s is not same .", pack, org, ocpversion))
					tnf.ClaimFilePrintf("Operator %s (organization %s) failed to be certified because of the wrong version of the operator or the ocp version %s is not same..", pack, org, ocpversion)
				} else {
					log.Info(fmt.Sprintf("Operator %s (organization %s) certified OK.", pack, org))
				}
			} else {
				tnf.ClaimFilePrintf("Operator %s is not a certified (needs to be part of the operator-certified organization in the catalog)", op.Packag)
			}
		}
		if testFailed {
			ginkgo.Skip("At least one  operator was not certified to run on this version of openshift. Check Claim.json file for details.")
		}
	})
}

func GetOcpVersion() string {
	ocCmd := ocpVersionCommand
	ocVersion := execCommandOutput(ocCmd)
	nums := strings.Split(strings.ReplaceAll(ocVersion, "\"", ""), ".")
	ocVersion = nums[0] + "." + nums[1]
	return ocVersion
}
