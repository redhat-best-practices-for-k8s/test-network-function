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
	mandatoryNumArgs = 1
)

func parseArgs() (*ping.Ping, time.Duration) {
	timeout := flag.Int("t", 2, "Timeout in seconds")
	count := flag.Int("c", 1, "Number of requests to send")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-t timeout] [-c count] host\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(incorrectUsageExitCode)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < mandatoryNumArgs {
		flag.Usage()
	}
	timeoutDuration := time.Duration(*timeout) * time.Second
	return ping.NewPing(timeoutDuration, args[0], *count), timeoutDuration
}

// Execute a ping test with exit code 0 on success, 1 on failure, 2 on error.
// Print interaction with the controlled subprocess which implements the test.
// Optionally log dialogue with the controlled subprocess to file.
func main() {
	result := tnf.ERROR
	pingReel, timeoutDuration := parseArgs()
	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner
	context, err := interactive.SpawnShell(&spawner, timeoutDuration, expect.Verbose(true))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(tnf.ExitCodeMap[result])
	}
	tester, err := tnf.NewTest(context.GetExpecter(), pingReel, []reel.Handler{pingReel}, context.GetErrorChannel())

	if err == nil {
		result, _ = tester.Run()
	} else {
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(tnf.ExitCodeMap[result])
}
