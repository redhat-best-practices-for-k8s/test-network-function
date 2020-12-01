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
	"fmt"
	"os"

	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/identifier"
)

const (
	// fatalErrorExitCode is the Unix return code used when there is a fatal error generating the catalog.
	fatalErrorExitCode = 1
)

// main generates a JSON formatted version of the test catalog.
func main() {
	contents, err := json.MarshalIndent(identifier.Catalog, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("could not generate the test catalog: %s", err))
		os.Exit(fatalErrorExitCode)
	}
	fmt.Print(string(contents))
}
