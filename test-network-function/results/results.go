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

package results

import (
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
	"github.com/test-network-function/test-network-function/pkg/junit"
)

var results = map[claim.Identifier][]claim.Result{}

// RecordResult is a hook provided to save aspects of the ginkgo.GinkgoTestDescription for a given claim.Identifier.
// Multiple results for a given identifier are aggregated as an array under the same key.
func RecordResult(identifier claim.Identifier) {
	testContext := ginkgo.CurrentGinkgoTestDescription()
	results[identifier] = append(results[identifier], claim.Result{
		Duration:      int(testContext.Duration.Nanoseconds()),
		Filename:      testContext.FileName,
		IsMeasurement: testContext.IsMeasurement,
		LineNumber:    testContext.LineNumber,
		TestText:      testContext.FullTestText,
	})
}

// GetReconciledResults is a function added to aggregate a Claim's results.  Due to the limitations of
// test-network-function-claim's Go Client, results are generalized to map[string]interface{}.  This method is needed
// to take the results gleaned from JUnit output, and to combine them with the contexts built up by subsequent calls to
// RecordResult.  The combination of the two forms a Claim's results.
func GetReconciledResults(testResults map[string]junit.TestResult) map[string]interface{} {
	resultMap := make(map[string]interface{})
	for key, vals := range results {
		// JSON cannot handle complex key types, so this flattens the complex key into a string format.
		strKey := fmt.Sprintf("{\"url\":\"%s\",\"version\":\"%s\"}", key.Url, key.Version)
		// initializes the result map, if necessary
		if _, ok := resultMap[strKey]; !ok {
			resultMap[strKey] = make([]claim.Result, 0)
		}
		// a codec which correlates claim.Result, JUnit results (testResults), and builds up the map
		// of claim's results.
		for _, val := range vals {
			val.Passed = testResults[val.TestText].Passed
			testFailReason := testResults[val.TestText].FailureReason
			val.FailureReason = testFailReason
			resultMap[strKey] = append(resultMap[strKey].([]claim.Result), val)
		}
	}
	return resultMap
}
