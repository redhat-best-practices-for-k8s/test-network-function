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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/test-network-function/test-network-function/pkg/gradetool"
)

const (
	flagResultsPath = "results"
	flagPolicyPath  = "policy"
	flagOutputPath  = "o"
)

func main() {
	resultsPath := flag.String(flagResultsPath, "", "Path to the input test results file")
	policyPath := flag.String(flagPolicyPath, "", "Path to the input policy file")
	outputPath := flag.String(flagOutputPath, "", "Path to the output file")
	flag.Parse()
	if resultsPath == nil || *resultsPath == "" {
		flag.Usage()
		return
	}
	if policyPath == nil || *policyPath == "" {
		flag.Usage()
		return
	}
	if outputPath == nil || *outputPath == "" {
		flag.Usage()
		return
	}

	err := gradetool.GenerateGrade(*resultsPath, *policyPath, *outputPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
