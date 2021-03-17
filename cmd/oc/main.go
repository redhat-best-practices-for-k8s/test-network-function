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
	"flag"
	"fmt"
	"os"
	"time"

	expect "github.com/ryandgoulding/goexpect"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ping"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	// incorrectUsageExitCode is the Unix return code used when the supplied arguments are inappropriate.
	incorrectUsageExitCode = 2

	// mandatoryNumArgs is the number of positional arguments required.
	mandatoryNumArgs = 3
)

func parseArgs() (*interactive.Oc, <-chan error, string, time.Duration, error) { //nolint:gocritic //permit unnamed return values
	timeout := flag.Int("t", 2, "Timeout in seconds")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-t timeout] pod container targetIpAddress ?oc-exec-opt ... oc-exec-opt?\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(incorrectUsageExitCode)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < mandatoryNumArgs {
		flag.Usage()
	}

	timeoutDuration := time.Duration(*timeout) * time.Second
	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner
	oc, ch, err := interactive.SpawnOc(&spawner, args[0], args[1], "default", timeoutDuration, expect.Verbose(true))
	return oc, ch, args[2], timeoutDuration, err
}

// Execute an OpenShift shell with exit code 0 on success, 1 on failure, 2 on error.
// Print interaction with the controlled subprocess which implements the session.
// Optionally log dialogue with the controlled subprocess to file.
// By default, read command lines to execute from stdin.
// Alternatively, read each input line as a JSON test configuration to execute.
func main() {
	result := tnf.ERROR
	oc, ch, targetIPAddress, timeoutDuration, err := parseArgs()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(tnf.ExitCodeMap[result])
	}

	request := ping.NewPing(timeoutDuration, targetIPAddress, 5)
	chain := []reel.Handler{request}
	test, err := tnf.NewTest(oc.GetExpecter(), request, chain, ch)

	if err == nil {
		result, err = test.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(tnf.ExitCodeMap[result])
}
