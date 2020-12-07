// Copyright (C) 2020 Red Hat, Inc.
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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/identifier"
	"github.com/spf13/cobra"
)

const (
	// fatalErrorExitCode is the Unix return code used when there is a fatal error generating the catalog.
	fatalErrorExitCode = 1
)

var (
	// rootCmd is the root of the "catalog" CLI program.
	rootCmd = &cobra.Command{
		Use:   "catalog",
		Short: "A CLI for creating the test catalog.",
	}

	// generateCmd is the root of the "catalog generate" CLI program.
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generates the test catalog",
	}

	// jsonGenerateCmd is used to generate a JSON formatted catalog to stdout.
	jsonGenerateCmd = &cobra.Command{
		Use:   "json",
		Short: "Generates the test catalog in JSON format.",
		Run:   runGenerateJSONCmd,
	}

	// markdownGenerateCmd is used to generate a markdown formatted catalog to stdout.
	markdownGenerateCmd = &cobra.Command{
		Use:   "markdown",
		Short: "Generates the test catalog in markdown format.",
		Run:   runGenerateMarkdownCmd,
	}
)

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
		n += len(elems[i])
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

// runGenerateMarkdownCmd generates a markdown test catalog.
func runGenerateMarkdownCmd(_ *cobra.Command, _ []string) {
	fmt.Println("# `tnf.Test` Catalog")
	fmt.Println()
	fmt.Println("A number of `tnf.Test` implementations are included out of the box.  This is a summary of the available implementations:")
	for _, catalogEntry := range identifier.Catalog {
		fmt.Fprintf(os.Stdout, "## %s", catalogEntry.Identifier.URL)
		fmt.Println()
		fmt.Println("Property|Description")
		fmt.Println("---|---")
		fmt.Fprintf(os.Stdout, "Version|%s\n", catalogEntry.Identifier.SemanticVersion)
		fmt.Fprintf(os.Stdout, "Description|%s\n", catalogEntry.Description)
		fmt.Fprintf(os.Stdout, "Result Type|%s\n", catalogEntry.Type)
		fmt.Fprintf(os.Stdout, "Intrusive|%t\n", catalogEntry.IntrusionSettings.ModifiesSystem)
		fmt.Fprintf(os.Stdout, "Modifications Persist After Test|%t\n", catalogEntry.IntrusionSettings.ModificationIsPersistent)
		fmt.Fprintf(os.Stdout, "Runtime Binaries Required|%s\n", cmdJoin(catalogEntry.BinaryDependencies, ", "))
		fmt.Println()
	}
}

// runGenerateJSONCmd generates a JSON test catalog.
func runGenerateJSONCmd(_ *cobra.Command, _ []string) {
	contents, err := json.MarshalIndent(identifier.Catalog, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("could not generate the test catalog: %s", err))
		os.Exit(fatalErrorExitCode)
	}
	fmt.Print(string(contents))
}

// Execute executes the "catalog" CLI.
func Execute() error {
	generateCmd.AddCommand(jsonGenerateCmd, markdownGenerateCmd)
	rootCmd.AddCommand(generateCmd)
	return rootCmd.Execute()
}
