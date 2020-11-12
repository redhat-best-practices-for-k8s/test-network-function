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

func parseArgs() (*ping.Ping, time.Duration) {
	timeout := flag.Int("t", 2, "Timeout in seconds")
	count := flag.Int("c", 1, "Number of requests to send")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-d logfile] [-t timeout] [-c count] host\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(tnf.ERROR)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
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
	ping, timeoutDuration := parseArgs()
	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner
	var err error
	var context *interactive.Context
	context, err = interactive.SpawnShell(&spawner, timeoutDuration, expect.Verbose(true))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(result)
	}
	printer := reel.NewPrinter("")
	tester, err := tnf.NewTest(context.GetExpecter(), ping, []reel.Handler{printer, ping}, context.GetErrorChannel())

	if err == nil {
		result, err = tester.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr,"Fatal Error Running Test: %v\n", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(result)
}
