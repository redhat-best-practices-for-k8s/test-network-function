package main

import (
	"flag"
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"os"
)

func parseArgs() (string, *tnf.IpAddr) {
	logfile := flag.String("d", "", "Filename to capture expect dialogue to")
	timeout := flag.Int("t", 2, "Timeout in seconds")
	device := flag.String("i", "eth0", "Interface Device")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-d logfile] [-t timeout] [-d device]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(tnf.ERROR)
	}
	flag.Parse()
	//args := flag.Args()
	//if len(args) == 0 {
	//	flag.Usage()
	//}
	return *logfile, tnf.NewIpAddr(*timeout, *device)
}

// Execute a ipaddr test with exit code 0 on success, 1 on failure, 2 on error.
// Print interaction with the controlled subprocess which implements the test.
// Optionally log dialogue with the controlled subprocess to file.
func main() {
	result := tnf.ERROR
	logfile, ipaddr := parseArgs()
	printer := reel.NewPrinter("")
	test, err := tnf.NewTest(logfile, ipaddr, []reel.Handler{printer, ipaddr})
	if err == nil {
		result, err = test.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(ipaddr.GetAddr())
	os.Exit(result)
}
