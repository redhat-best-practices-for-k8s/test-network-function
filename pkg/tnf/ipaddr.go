package tnf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"regexp"
)

const ipdone string = `\s+inet ((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))`

// A ping test implemented using command line tool `ping`.
type IpAddr struct {
	result  int
	timeout int
	device string
	addr   string
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
		Expect: []string{ipdone},
		Timeout: i.timeout,
	}
}

func (i *IpAddr) ReelMatch(pattern string, before string, match string) *reel.Step {
	re := regexp.MustCompile(ipdone)
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
		Expect: []string{ipdone},
	}
}

func (i *IpAddr) ReelEof() {
	// empty
}

func IpAddrCmd(device string) [] string {
	return []string{"ip", "address", "show", "dev", device}
}

func NewIpAddr(timeout int, device string) *IpAddr {
	return &IpAddr{
		result: ERROR,
		timeout: timeout,
		device: device,
		args: IpAddrCmd(device),
	}
}