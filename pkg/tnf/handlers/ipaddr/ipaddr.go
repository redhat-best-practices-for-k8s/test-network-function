// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package ipaddr

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/dependencies"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

// IPAddr provides an ip addr test implemented using command line tool `ip addr`.
type IPAddr struct {
	result  int
	timeout time.Duration
	args    []string
	// The ipv4 address for a given device if the Handler matches.
	ipv4Address string
}

const (
	// DeviceDoesNotExistRegex matches `ip addr` output when the given device does not exist.
	DeviceDoesNotExistRegex = `(?m)Device \"(\w+)\" does not exist.$`
	// SuccessfulOutputRegex matches `ip addr` output for a given device, and provides grouping to extract the associated Ipv4 address.
	SuccessfulOutputRegex = `(?m)^\s+inet ((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))`
)

var (
	// ipAddrCommand is the skeleton command used to get ip address information for a device.
	ipAddrCommand = fmt.Sprintf("%s addr show dev", dependencies.IPBinaryName)
)

// Args returns the command line args for the test.
func (i *IPAddr) Args() []string {
	return i.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (i *IPAddr) GetIdentifier() identifier.Identifier {
	return identifier.IPAddrIdentifier
}

// Timeout return the timeout for the test.
func (i *IPAddr) Timeout() time.Duration {
	return i.timeout
}

// Result returns the test result.
func (i *IPAddr) Result() int {
	return i.result
}

// ReelFirst returns a step which expects an ip summary for the given device.
func (i *IPAddr) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{SuccessfulOutputRegex, DeviceDoesNotExistRegex},
		Timeout: i.timeout,
	}
}

// ReelMatch parses the ip addr output and set the test result on match.
// Returns no step; the test is complete.
func (i *IPAddr) ReelMatch(pattern, _, match string) *reel.Step {
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

// ReelTimeout does nothing;  no intervention is needed for `ip addr` timeout.
func (i *IPAddr) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no intervention is needed for `ip addr` EOF.
func (i *IPAddr) ReelEOF() {
}

// GetIPv4Address returns the extracted IPv4 address for the given device (interface).
func (i *IPAddr) GetIPv4Address() string {
	return i.ipv4Address
}

func ipAddrCmd(dev string) []string {
	return strings.Split(fmt.Sprintf("%s %s", ipAddrCommand, dev), " ")
}

// NewIPAddr creates a new `ip addr` test for the given device.
func NewIPAddr(timeout time.Duration, device string) *IPAddr {
	return &IPAddr{result: tnf.ERROR, timeout: timeout, args: ipAddrCmd(device)}
}
