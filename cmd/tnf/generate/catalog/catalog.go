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

package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/test-network-function/test-network-function-claim/pkg/claim"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"

	"github.com/spf13/cobra"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
)

const (
	// backtickOffset is the number of extra characters required to enclose output in backticks.
	backtickOffset = 2

	// introMDFilename is the name of the file that contains the introductory text for CATALOG.md.
	introMDFilename = "INTRO.md"

	// tccFilename is the name of the file that contains the test case catalog section introductory text for CATALOG.md.
	tccFilename = "TEST_CASE_CATALOG.md"

	// tccbbFilename is the name of the file that contains the test case catalog building blocks section introductory
	// text for CATALOG.md.
	tccbbFilename = "TEST_CASE_BUILDING_BLOCKS_CATALOG.md"
)

var (
	// introMDFile is the path to the file that contains the test case catalog section introductory text for CATALOG.md.
	introMDFile = path.Join(mdDirectory, introMDFilename)

	// mdDirectory is the path to the directory of files that contain static text for CATALOG.md.
	mdDirectory = path.Join("cmd", "tnf", "generate", "catalog")

	// tccFile is the path to the file that contains the test case catalog section introductory text for CATALOG.md.
	tccFile = path.Join(mdDirectory, tccFilename)

	// tccbbFile is the path to the file that contains the test case catalog building blocks section introductory text
	// for CATALOG.md
	tccbbFile = path.Join(mdDirectory, tccbbFilename)

	// generateCmd is the root of the "catalog generate" CLI program.
	generateCmd = &cobra.Command{
		Use:   "catalog",
		Short: "Generates the test catalog",
	}

	// jsonGenerateCmd is used to generate a JSON formatted catalog to stdout.
	jsonGenerateCmd = &cobra.Command{
		Use:   "json",
		Short: "Generates the test catalog in JSON format.",
		RunE:  runGenerateJSONCmd,
	}

	// markdownGenerateCmd is used to generate a markdown formatted catalog to stdout.
	markdownGenerateCmd = &cobra.Command{
		Use:   "markdown",
		Short: "Generates the test catalog in markdown format.",
		RunE:  runGenerateMarkdownCmd,
	}
)

type CatalogElement struct {
	testName   string
	identifier claim.Identifier // {url and version}
}

// cmdJoin is a utility method abstracted from strings.Join which shims in better formatting for markdown files.
func cmdJoin(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return "`" + elems[0] + "`"
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		// backtickOffset is used to track the extra length required by enclosing output commands in backticks.
		n += len(elems[i]) + backtickOffset
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString("`" + elems[0] + "`")
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString("`" + s + "`")
	}
	return b.String()
}

// emitTextFromFile is a utility method to stream file contents to stdout.  This allows more natural specification of
// the non-dynamic aspects of CATALOG.md.
func emitTextFromFile(filename string) error {
	text, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	fmt.Print(string(text))
	return nil
}

// it extracts the suite name and test name from an url such as
// http://test-network-function.com/tests-case/SuitName/TestName

func getSuiteAndTestFromUrl(url string, base_domain string) []string {
	result := strings.Split(url, base_domain)
	// len 2, because returns [0] before base_domain and [1] after base_domain
	if len(result) > 2 {
		fmt.Fprintf(os.Stderr, "Identifier Url not valid\n")
		return nil
	}

	suite_test := strings.Split(result[1], "/")
	return suite_test
}

// it extracts the test name and test name from an url such as
// http://test-network-function.com/tests/TestName

func getTestFromUrl(url string, base_domain string) string {
	result := strings.Split(url, base_domain)
	// len 2, because returns [0] before base_domain and [1] after base_domain
	if len(result) > 2 {
		fmt.Fprintf(os.Stderr, "Identifier Url not valid\n")
		return ""
	}
	test_name := result[1]
	return test_name
}

// it turns a list of identifiers
// { url, version  }
// and takes urls like http://test-network-function.com/testcases/SuiteName/TestName
// to build a more structured catalogue
// {
//     suiteNameA: [testName [identifiers], testName2 [identifiers] ]
//     suiteNameB: [testName3 [identifiers], testName4 [identifiers] ]
// }
func createPrintableCatalogFromIdentifiers(keys []claim.Identifier) map[string][]CatalogElement {
	base_domain := identifiers.TestIDBaseDomain + "/"

	catalog := make(map[string][]CatalogElement)
	// we need the list of suite's names
	var element CatalogElement

	for _, i := range keys {
		suite_test := getSuiteAndTestFromUrl(i.Url, base_domain)
		if suite_test == nil {
			return nil
		}
		suite_name := suite_test[0]
		test_name := suite_test[1]
		element.testName = test_name
		element.identifier = i
		catalog[suite_name] = append(catalog[suite_name], element)
	}
	return catalog
}

// it turns a list of urls http://test-network-function.com/tests/TestName
// to build a more structured catalogue
// {
//     0: [testName [identifiers], testName2 [identifiers] ]
//     1: [testName3 [identifiers], testName4 [identifiers] ]
// }
func createPrintableCatalogFromUrls(urls []string) map[int][]CatalogElement {
	base_domain := identifier.TestIDBaseDomain + "/"

	catalog := make(map[int][]CatalogElement)
	// we need the list of suite's names
	var element CatalogElement
	var c int = 0

	for _, url := range urls {
		fmt.Println(url)
		test_name := getTestFromUrl(url, base_domain)
		if test_name == "" {
			return nil
		}
		element.testName = test_name
		element.identifier = claim.Identifier{Url: url}
		catalog[c] = append(catalog[c], element)
		c++
	}
	return catalog
}

func getSuitesFromIdentifiers(keys []claim.Identifier) []string {
	base_domain := identifiers.TestIDBaseDomain + "/"
	var suites []string

	for _, i := range keys {
		suite_test := getSuiteAndTestFromUrl(i.Url, base_domain)
		suite_name := suite_test[0]
		suites = append(suites, suite_name)
	}
	return Unique(suites)
}

func Unique(slice []string) []string {
	// create a map with all the values as key
	uniqMap := make(map[string]struct{})
	for _, v := range slice {
		uniqMap[v] = struct{}{}
	}

	// turn the map keys into a slice
	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	return uniqSlice
}

// outputTestCases outputs the Markdown representation for test cases from the catalog to stdout.
func outputTestCases() {
	// Building a separate data structure to store the key order for the map
	keys := make([]claim.Identifier, 0, len(identifiers.Catalog))
	for k := range identifiers.Catalog {
		keys = append(keys, k)
	}

	// Sorting the map by identifier URL
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Url < keys[j].Url
	})

	catalog := createPrintableCatalogFromIdentifiers(keys)
	if catalog == nil {
		return
	}
	// we need the list of suite's names
	suites := getSuitesFromIdentifiers(keys)

	// Iterating the map by test and suite names
	for _, suite := range suites {
		fmt.Println()
		fmt.Fprintf(os.Stdout, "### %s\n", suite)
		fmt.Println()
		for _, k := range catalog[suite] {
			fmt.Println()
			fmt.Fprintf(os.Stdout, "#### %s\n", k.testName)
			fmt.Println()

			fmt.Println("Property|Description")
			fmt.Println("---|---")
			fmt.Fprintf(os.Stdout, "Test Name|%s\n", k.testName)
			fmt.Fprintf(os.Stdout, "Url|%s\n", k.identifier.Url)
			fmt.Fprintf(os.Stdout, "Version|%s\n", k.identifier.Version)
			fmt.Fprintf(os.Stdout, "Description|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Description, "\n", " "))
			fmt.Fprintf(os.Stdout, "Result Type|%s\n", identifiers.Catalog[k.identifier].Type)
			fmt.Fprintf(os.Stdout, "Suggested Remediation|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].Remediation, "\n", " "))
			fmt.Fprintf(os.Stdout, "Best Practice Reference|%s\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].BestPracticeReference, "\n", " "))
		}
	}
	fmt.Println()
	fmt.Println()
}

// outputTestCaseBuildingBlocks outputs the Markdown representation for the test case building blocks from the catalog
// to stdout.
func outputTestCaseBuildingBlocks() {
	// Building a separate data structure to store the key order for the map
	keys := make([]string, 0, len(identifier.Catalog))
	for k := range identifier.Catalog {
		keys = append(keys, k)
	}

	// Sorting the map by identifier URL
	sort.Strings(keys)

	catalog := createPrintableCatalogFromUrls(keys)
	if catalog == nil {
		return
	}

	for i := 0; i < len(catalog); i++ {
		for _, k := range catalog[i] {

			fmt.Println()
			fmt.Fprintf(os.Stdout, "### %s\n", k.testName)
			fmt.Println()

			fmt.Println()
			fmt.Println("Property|Description")
			fmt.Println("---|---")
			fmt.Fprintf(os.Stdout, "Test Name|%s\n", k.testName)
			fmt.Fprintf(os.Stdout, "Url|%s\n", identifier.Catalog[k.identifier.Url].Identifier.URL)
			fmt.Fprintf(os.Stdout, "Version|%s\n", identifier.Catalog[k.identifier.Url].Identifier.SemanticVersion)
			fmt.Fprintf(os.Stdout, "Description|%s\n", identifier.Catalog[k.identifier.Url].Description)
			fmt.Fprintf(os.Stdout, "Result Type|%s\n", identifier.Catalog[k.identifier.Url].Type)
			fmt.Fprintf(os.Stdout, "Intrusive|%t\n", identifier.Catalog[k.identifier.Url].IntrusionSettings.ModifiesSystem)
			fmt.Fprintf(os.Stdout, "Modifications Persist After Test|%t\n", identifier.Catalog[k.identifier.Url].IntrusionSettings.ModificationIsPersistent)
			fmt.Fprintf(os.Stdout, "Runtime Binaries Required|%s\n", cmdJoin(identifier.Catalog[k.identifier.Url].BinaryDependencies, ", "))
			fmt.Println()
		}
	}
}

// runGenerateMarkdownCmd generates a markdown test catalog.
func runGenerateMarkdownCmd(_ *cobra.Command, _ []string) error {
	// static introductory generation
	if err := emitTextFromFile(introMDFile); err != nil {
		return err
	}
	if err := emitTextFromFile(tccFile); err != nil {
		return err
	}

	// process the test cases
	outputTestCases()

	// static generation of test case building blocks section introduction.
	if err := emitTextFromFile(tccbbFile); err != nil {
		return err
	}

	// process the test case building blocks
	outputTestCaseBuildingBlocks()
	return nil
}

// runGenerateJSONCmd generates a JSON test catalog.
func runGenerateJSONCmd(_ *cobra.Command, _ []string) error {
	contents, err := json.MarshalIndent(identifier.Catalog, "", "  ")
	if err != nil {
		return err
	}
	fmt.Print(string(contents))
	return nil
}

// Execute executes the "catalog" CLI.
func NewCommand() *cobra.Command {
	generateCmd.AddCommand(jsonGenerateCmd, markdownGenerateCmd)
	return generateCmd
}
