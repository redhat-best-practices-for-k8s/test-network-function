package ipaddr

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"regexp"
	"strings"
	"time"
)

// An ip addr test implemented using command line tool `ip`.
type IpAddr struct {
	result  int
	timeout time.Duration
	args    []string
	// The ipv4 address for a given device if the Handler matches.
	ipv4Address string
}

const (
	ipAddrCommand           = "ip addr show dev"
	DeviceDoesNotExistRegex = `(?m)Device \"(\w+)\" does not exist.$`
	SuccessfulOutputRegex   = `(?m)^\s+inet ((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))`
)

// Return the command line args for the test.
func (i *IpAddr) Args() []string {
	return i.args
}

// Return the timeout in seconds for the test.
func (i *IpAddr) Timeout() time.Duration {
	return i.timeout
}

// Return the test result.
func (i *IpAddr) Result() int {
	return i.result
}

// Return a step which expects an ip summary for the given device.
func (i *IpAddr) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{SuccessfulOutputRegex, DeviceDoesNotExistRegex},
		Timeout: i.timeout,
	}
}

// On match, parse the ip addr output and set the test result.
// Returns no step; the test is complete.
func (i *IpAddr) ReelMatch(pattern string, _ string, match string) *reel.Step {
	if pattern == DeviceDoesNotExistRegex {
		i.result = tnf.ERROR
		return nil
	}
	re := regexp.MustCompile(SuccessfulOutputRegex)
	matched := re.FindStringSubmatch(match)
	if matched != nil {
		i.ipv4Address = matched[1]
		i.result = tnf.SUCCESS
	}
	return nil
}

// On timeout, return a step which kills the ping test by sending it ^C.
func (i *IpAddr) ReelTimeout() *reel.Step {
	return nil
}

// On eof, take no action.
func (i *IpAddr) ReelEof() {
}

func (i *IpAddr) GetIpv4Address() string {
	return i.ipv4Address
}

func ipAddrCmd(dev string) []string {
	return strings.Split(fmt.Sprintf("%s %s", ipAddrCommand, dev), " ")
}

// Create a new `Ping` test which pings `hosts` with `count` requests, or
// indefinitely if `count` is not positive, and executes within `timeout`
// seconds.
func NewIpAddr(timeout time.Duration, dev string) *IpAddr {
	return &IpAddr{result: tnf.ERROR, timeout: timeout, args: ipAddrCmd(dev)}
}
