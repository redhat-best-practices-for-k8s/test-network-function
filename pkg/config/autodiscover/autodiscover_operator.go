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
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	operatorLabelName = "operator"
)

var (
	operatorTestsAnnotationName    = buildAnnotationName("operator_tests")
	subscriptionNameAnnotationName = buildAnnotationName("subscription_name")
)

// BuildOperatorConfig builds a `[]configsections.Operator` from the current state of the cluster,
// using labels and annotations to populate the data.
func BuildOperatorConfig() (operatorsToTest []configsections.Operator) {
	csvs, err := GetCSVsByLabel(operatorLabelName, AnyLabelValue)
	if err != nil {
		log.Fatalf("found no CSVs to test while 'operator' spec enabled: %s", err)
	}
	for i := range csvs.Items {
		operatorsToTest = append(operatorsToTest, BuildOperatorFromCSVResource(&csvs.Items[i]))
	}
	return operatorsToTest
}

// BuildOperatorFromCSVResource builds a single `configsections.Operator` from a CSVResource
func BuildOperatorFromCSVResource(csv *CSVResource) (op configsections.Operator) {
	var err error
	op.Name = csv.Metadata.Name
	op.Namespace = csv.Metadata.Namespace

	var tests []string
	err = csv.GetAnnotationValue(operatorTestsAnnotationName, &tests)
	if err != nil {
		log.Warnf("unable to extract tests from annotation on '%s/%s' (error: %s). Attempting to fallback to all tests", op.Namespace, op.Name, err)
		op.Tests = getConfiguredOperatorTests()
	} else {
		op.Tests = tests
	}

	var subscriptionName string
	err = csv.GetAnnotationValue(subscriptionNameAnnotationName, &subscriptionName)
	if err != nil {
		log.Warnf("unable to get a subscription name annotation from CSV %s (%s), the CSV name will be used", csv.Metadata.Name, err)
	}
	op.SubscriptionName = subscriptionName

	return op
}

// getConfiguredOperatorTests loads the `configuredTestFile` used by the `operator` specs and extracts
// the names of test groups from it.
func getConfiguredOperatorTests() (opTests []string) {
	configuredTests, err := testcases.LoadConfiguredTestFile(configuredTestFile)
	if err != nil {
		log.Errorf("failed to load %s, continuing with no tests", configuredTestFile)
		return []string{}
	}
	for _, configuredTest := range configuredTests.OperatorTest {
		opTests = append(opTests, configuredTest.Name)
	}
	log.WithField("opTests", opTests).Infof("got all tests from %s.", configuredTestFile)
	return opTests
}
