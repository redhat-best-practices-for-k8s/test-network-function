package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/reel"
	"github.com/redhat-nfvpe/test-network-function/tnf"
	"os"
)

func tester() (string, reel.Handler, *tnf.Ssh) {
	logfile := flag.String("d", "", "Filename to capture expect dialogue to")
	timeout := flag.Int("t", 2, "Timeout in seconds")
	testers := flag.Bool("T", false, "Feed tests as JSON from stdin")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-d logfile] [-t timeout] [-T] prompt host ?ssh-opt .. ssh-opt?\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(tnf.ERROR)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
	}
	var feeder reel.Handler
	if *testers {
		feeder = tnf.NewTestFeeder(*timeout, args[0], bufio.NewScanner(os.Stdin))
	} else {
		feeder = reel.NewLineFeeder(*timeout, args[0], bufio.NewScanner(os.Stdin))
	}
	ssh := tnf.NewSsh(*timeout, args[0], args[1], args[2:])
	return *logfile, feeder, ssh
}

func main() {
	result := tnf.ERROR
	logfile, feeder, ssh := tester()
	printer := reel.NewPrinter(" \r\n")
	test, err := tnf.NewTest(logfile, ssh, []reel.Handler{printer, feeder, ssh})
	if err == nil {
		result, err = test.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(result)
}
