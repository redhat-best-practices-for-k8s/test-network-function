package tnf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"regexp"
)

const Ipdone string = `\s+inet ((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))`

// Extracts IP Address info for a remote container.
type IpAddr struct {
	result  int
	timeout int
	// The network interface in question.
	device string
	// Stores the IPv4 address for device.
	addr string
	args []string
}

func (i *IpAddr) GetAddr() string {
	return i.addr
}

func (i *IpAddr) Args() []string {
	return i.args
}

func (i *IpAddr) Device() string {
	return i.device
}

func (i *IpAddr) Timeout() int {
	return i.timeout
}

func (i *IpAddr) Result() int {
	return i.result
}

func (i *IpAddr) ReelFirst() *reel.Step {
	return &reel.Step{
		// TODO
		Expect:  []string{Ipdone},
		Timeout: i.timeout,
	}
}

func (i *IpAddr) ReelMatch(pattern string, before string, match string) *reel.Step {
	re := regexp.MustCompile(Ipdone)
	matched := re.FindStringSubmatch(match)
	if matched != nil {
		i.addr = matched[1]
		i.result = SUCCESS
	}
	return nil
}

func (i *IpAddr) ReelTimeout() *reel.Step {
	return &reel.Step{
		Execute: reel.CTRL_C,
		Expect:  []string{Ipdone},
	}
}

func (i *IpAddr) ReelEof() {
	// empty
}

func IpAddrCmd(pod, container, device string, opts []string) []string {
	args := []string{"oc", "exec", pod, "-c", container}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	return append(args, "--", "ip", "address", "show", "dev", device)
}

func NewIpAddr(timeout int, pod, container, device string, opts []string) *IpAddr {
	return &IpAddr{
		result:  ERROR,
		timeout: timeout,
		device:  device,
		args:    IpAddrCmd(pod, container, device, opts),
	}
}
