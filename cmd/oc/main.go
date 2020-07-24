package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"os"
)

func parseArgs() (string, reel.Handler, *tnf.Oc) {
	logfile := flag.String("d", "", "Filename to capture expect dialogue to")
	timeout := flag.Int("t", 2, "Timeout in seconds")
	feed := flag.String("f", "", "Feed 'tests' (JSON configurations) or 'lines' from stdin")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-d logfile] [-t timeout] [-f 'lines'|'tests'] pod ?oc-exec-opt .. oc-exec-opt?\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(tnf.ERROR)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
	}
	var feeder reel.Handler
	switch *feed {
	case "tests":
		feeder = tnf.NewTestFeeder(*timeout, tnf.OcPrompt, bufio.NewScanner(os.Stdin))
	case "lines":
		feeder = reel.NewLineFeeder(*timeout, tnf.OcPrompt, bufio.NewScanner(os.Stdin))
	default:
		feeder = nil
	}
	oc := tnf.NewOc(*timeout, args[0], args[1:])
	return *logfile, feeder, oc
}

// Execute an OpenShift shell with exit code 0 on success, 1 on failure, 2 on error.
// Print interaction with the controlled subprocess which implements the session.
// Optionally log dialogue with the controlled subprocess to file.
// By default, read command lines to execute from stdin.
// Alternatively, read each input line as a JSON test configuration to execute.
func main() {
	result := tnf.ERROR
	logfile, feeder, oc := parseArgs()
	printer := reel.NewPrinter(" \r\n")
	var chain []reel.Handler
	if feeder != nil {
		chain = []reel.Handler{printer, feeder, oc}
	} else {
		chain = []reel.Handler{printer, oc}
	}
	test, err := tnf.NewTest(logfile, oc, chain)
	if err == nil {
		result, err = test.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(result)
}
