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
)

var (
	subscriptionCommand = "oc get subscriptions.operators.coreos.com -A -ogo-template='{{range .items}}{{.spec.source}}_{{.status.currentCSV}},{{end}}'"
	ocpVersionCommand   = "oc version -o json | jq '.openshiftVersion'"

	execCommandOutput = func(command string) string {
		return utils.ExecuteCommandAndValidate(command, apiRequestTimeout, interactive.GetContext(expectersVerboseModeEnabled), func() {
			log.Error("can't run command: ", command)
		})
	}

	certAPIClient api.CertAPIClient
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
		testOperatorCertificationStatus()
		testCSICertified(env)
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
	ocpversion := GetOcpVersion()
	return func() bool {
		return certAPIClient.IsOperatorCertified(organization, operatorName, ocpversion)
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

//nolint:dupl
func testContainerCertificationStatus() {
	// Query API for certification status of listed containers
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestContainerIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		env := configpkg.GetTestEnvironment()
		containersToQuery := env.Config.CertifiedContainerInfo

		if len(containersToQuery) == 0 {
			ginkgo.Skip("No containers to check configured in tnf_config.yml")
		}

		ginkgo.By(fmt.Sprintf("Getting certification status. Number of containers to check: %d", len(containersToQuery)))

		if len(containersToQuery) > 0 {
			certAPIClient = api.NewHTTPClient()
			failedContainers := []configsections.CertifiedContainerRequestInfo{}
			allContainersToQueryEmpty := true
			for _, c := range containersToQuery {
				if c.Name == "" || c.Repository == "" {
					tnf.ClaimFilePrintf("Container name = \"%s\" or repository = \"%s\" is missing, skipping this container to query", c.Name, c.Repository)
					continue
				}
				allContainersToQueryEmpty = false
				ginkgo.By(fmt.Sprintf("Container %s/%s should eventually be verified as certified", c.Repository, c.Name))
				isCertified := waitForCertificationRequestToSuccess(getContainerCertificationRequestFunction(c.Repository, c.Name), apiRequestTimeout)
				if !isCertified {
					tnf.ClaimFilePrintf("Container %s (repository %s) failed to be certified.", c.Name, c.Repository)
					failedContainers = append(failedContainers, c)
				} else {
					log.Info(fmt.Sprintf("Container %s (repository %s) certified OK.", c.Name, c.Repository))
				}
			}
			if allContainersToQueryEmpty {
				ginkgo.Skip("No containers to check because either container name or repository is empty for all containers in tnf_config.yml")
			}

			if n := len(failedContainers); n > 0 {
				log.Warnf("Containers that failed to be certified: %+v", failedContainers)
				ginkgo.Fail(fmt.Sprintf("%d containers failed to be certified.", n))
			}
		}
	})
}

//nolint:dupl
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

func testCSICertified(env *configpkg.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestCSIOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		csioperatorsToQuery := env.OperatorsUnderTest

		if len(csioperatorsToQuery) == 0 {
			ginkgo.Skip("No CSI operators to check configured ")
		}

		ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of CSI drivers to check: %d", len(csioperatorsToQuery)))

		//mapOperatorVersions := csimapping.GetOperatorVersions()
		//ocpVersion := GetOcpVersion()
		//operatorVersionMap, orgMap := GetOperatorVersionMap()
		testFailed := false
		for _, csi := range csioperatorsToQuery {
			pack := csi.Name
			org := csi.Org
			if org != "certified-operators" {
				isCertified := waitForCertificationRequestToSuccess(getOperatorCertificationRequestFunction(org, pack), apiRequestTimeout)
				if !isCertified {
					tnf.ClaimFilePrintf("Operator %s (organization %s) failed to be certified.", pack, org)
				} else {
					log.Info(fmt.Sprintf("Operator %s (organization %s) certified OK.", pack, org))
				}
			} else {
				tnf.ClaimFilePrintf("Driver %s is not a certified CSI driver (needs to be part of the operator-certified organization in the catalog) or csimapping.json needs to be updated", csi.Packag)
			}
		}
		if testFailed {
			ginkgo.Fail("At least one CSI operator was not certified to run on this version of openshift. Check Claim.json file for details.")
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
func GetOperatorVersionMap() (versionMap, orgMap map[string]string) {
	ocCmd := subscriptionCommand

	out := execCommandOutput(ocCmd)

	operatorVersionList := strings.Split(out, ",")
	versionMap = make(map[string]string)
	orgMap = make(map[string]string)

	for _, entry := range operatorVersionList {
		if entry == "" {
		} else {
			organizationVersion := strings.SplitN(entry, "_", 2) //nolint:gomnd // ok
			org := organizationVersion[0]
			nameVersion := strings.SplitN(organizationVersion[1], ".", 2) //nolint:gomnd // ok
			name := nameVersion[0]
			version := nameVersion[1]
			versionMap[name] = version
			orgMap[name] = org
		}
	}

	return versionMap, orgMap
}
