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

package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/test-network-function/test-network-function-claim/pkg/claim"
	"github.com/test-network-function/test-network-function/pkg/junit"

	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	argsLen              = 2
	claimAdd             = "claim-add"
	claimFilePermissions = 0644
)

var (
	// claim-add subcommand flag pointers
	// Adding a new choice for --claimfile of 'substring' and a new --reportFiles flag
	claimFileTextPtr   *string
	reportFilesTextPtr *string
)

func main() {
	// Subcommands
	claimAddCommand := flag.NewFlagSet("claim-add", flag.ExitOnError)

	// claim-add subcommand flag pointers
	// Adding a new choice for --claimfile of 'substring' and a new --reportFiles flag
	claimFileTextPtr = claimAddCommand.String("claimfile", "", "existing claim file. (Required)")
	reportFilesTextPtr = claimAddCommand.String("reportdir", "", "dir of JUnit XML reports. (Required)")

	// Verify that a subcommand has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if len(os.Args) < argsLen {
		log.Fatalf("claim-add subcommand is required")
	}

	// Switch on the subcommand
	// Parse the flags for appropriate FlagSet
	// FlagSet.Parse() requires a set of arguments to parse as input
	// os.Args[2:] will be all arguments starting after the subcommand at os.Args[1]
	switch os.Args[1] {
	case claimAdd:
		if err := claimAddCommand.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Error reading argument  %v", err)
		}
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if claimAddCommand.Parsed() {
		// Required Flags
		if *claimFileTextPtr == "" {
			claimAddCommand.PrintDefaults()
			os.Exit(1)
		}
		if *reportFilesTextPtr == "" {
			claimAddCommand.PrintDefaults()
			os.Exit(1)
		}
		claimUpdate()
	}
}

func claimUpdate() {
	fileUpdated := false
	dat, err := ioutil.ReadFile(*claimFileTextPtr)
	if err != nil {
		log.Fatalf("Error reading claim file :%v", err)
	}

	claimRoot := readClaim(&dat)
	junitMap := claimRoot.Claim.Results

	items, _ := ioutil.ReadDir(*reportFilesTextPtr)

	for _, item := range items {
		fileName := item.Name()
		extension := filepath.Ext(fileName)
		reportKeyName := fileName[0 : len(fileName)-len(extension)]

		if _, ok := junitMap[reportKeyName]; ok {
			log.Printf("Skipping: %s already exists in supplied `%s` claim file", reportKeyName, *claimFileTextPtr)
		} else {
			junitMap[reportKeyName], err = junit.ExportJUnitAsJSON(fmt.Sprintf("%s/%s", *reportFilesTextPtr, item.Name()))
			if err != nil {
				log.Fatalf("Error reading JUnit XML file into JSON: %v", err)
			}
			fileUpdated = true
		}
	}
	claimRoot.Claim.Results = junitMap
	payload, err := json.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate the claim: %v", err)
	}
	err = ioutil.WriteFile(*claimFileTextPtr, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
	}
	if fileUpdated {
		log.Printf("Claim file `%s` updated\n", *claimFileTextPtr)
	} else {
		log.Printf("No changes were applied to `%s`\n", *claimFileTextPtr)
	}
}

func readClaim(contents *[]byte) *claim.Root {
	var claimRoot claim.Root
	err := json.Unmarshal(*contents, &claimRoot)
	if err != nil {
		log.Fatalf("Error reading claim constents file into type: %v", err)
	}
	return &claimRoot
}
