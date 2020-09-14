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

func parseArgs() (*interactive.Context, string, time.Duration, error) {
	timeout := flag.Int("t", 2, "Timeout in seconds")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-t timeout] user host targetIpAddress\n", os.Args[0])
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
	context, err := interactive.SpawnSSH(&spawner, args[0], args[1], time.Duration(10), expect.Verbose(true))
	return context, args[2], timeoutDuration, err
}

// Execute a SSH session with exit code 0 on success, 1 on failure, 2 on error.
// Execute a ping to the target IP address and print interaction with the controlled subprocess.
func main() {
	result := tnf.ERROR
	context, targetIPAddress, timeoutDuration, err := parseArgs()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(result)
	}

	printer := reel.NewPrinter(" \r\n")
	request := ping.NewPing(timeoutDuration, targetIPAddress, 5)
	chain := []reel.Handler{printer, request}
	test, err := tnf.NewTest(context.GetExpecter(), request, chain, context.GetErrorChannel())

	if err == nil {
		result, err = test.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(result)
}
