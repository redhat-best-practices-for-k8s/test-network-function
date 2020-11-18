package main

import (
	"flag"
	"fmt"
	expect "github.com/google/goexpect"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/ping"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"os"
	"time"
)

func parseArgs() (*interactive.Oc, <-chan error, string, time.Duration, error) {
	timeout := flag.Int("t", 2, "Timeout in seconds")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-t timeout] pod container targetIpAddress ?oc-exec-opt ... oc-exec-opt?\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(tnf.ERROR)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 3 {
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
		os.Exit(result)
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
	os.Exit(result)
}
