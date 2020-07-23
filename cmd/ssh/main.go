package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"os"
)

func parseArgs() (string, reel.Handler, *tnf.Ssh) {
	logfile := flag.String("d", "", "Filename to capture expect dialogue to")
	timeout := flag.Int("t", 2, "Timeout in seconds")
	feed := flag.String("f", "", "Feed 'tests' (JSON configurations) or 'lines' from stdin")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-d logfile] [-t timeout] [-f 'lines'|'tests'] prompt host ?ssh-opt .. ssh-opt?\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(tnf.ERROR)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
	}
	var feeder reel.Handler
	switch *feed {
	case "tests":
		feeder = tnf.NewTestFeeder(*timeout, args[0], bufio.NewScanner(os.Stdin))
	case "lines":
		feeder = reel.NewLineFeeder(*timeout, args[0], bufio.NewScanner(os.Stdin))
	default:
		feeder = nil
	}
	ssh := tnf.NewSsh(*timeout, args[0], args[1], args[2:])
	return *logfile, feeder, ssh
}

// Execute a SSH session with exit code 0 on success, 1 on failure, 2 on error.
// Print interaction with the controlled subprocess which implements the session.
// Optionally log dialogue with the controlled subprocess to file.
// By default, read command lines to execute from stdin.
// Alternatively, read each input line as a JSON test configuration to execute.
func main() {
	result := tnf.ERROR
	logfile, feeder, ssh := parseArgs()
	printer := reel.NewPrinter(" \r\n")
	var chain []reel.Handler
	if feeder != nil {
		chain = []reel.Handler{printer, feeder, ssh}
	} else {
		chain = []reel.Handler{printer, ssh}
	}
	test, err := tnf.NewTest(logfile, ssh, chain)
	if err == nil {
		result, err = test.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(result)
}
