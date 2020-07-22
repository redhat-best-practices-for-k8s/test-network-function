package main

import (
	"flag"
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"os"
)

func tester() (string, *tnf.Ping) {
	logfile := flag.String("d", "", "Filename to capture expect dialogue to")
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
	return *logfile, tnf.NewPing(*timeout, args[0], *count)
}

func main() {
	result := tnf.ERROR
	logfile, ping := tester()
	printer := reel.NewPrinter("")
	test, err := tnf.NewTest(logfile, ping, []reel.Handler{printer, ping})
	if err == nil {
		result, err = test.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(result)
}
