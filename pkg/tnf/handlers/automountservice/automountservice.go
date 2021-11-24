// Copyright (C) 2020-2021 Red Hat, Inc.
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

package automountservice

import (
	"regexp"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/test-network-function/common"
)

const (
	TokenIsTrue  int = 1
	TokenIsFalse int = 2
	TokenNotSet  int = 3
)

// Automountservice is the reel handler struct.
type Automountservice struct {
	result         int
	podCheck       bool
	namespace      string
	serviceaccount string
	podname        string
	timeout        time.Duration
	args           []string
	token          int
}

const (
	saRegex = `(?m)"automountServiceAccountToken": (.+)`
	False   = "false,"
	True    = "true,"
)

// NewAutomountservice returns a new Automountservice handler struct.
func NewAutomountservice(options ...func(*Automountservice)) *Automountservice {
	as := &Automountservice{
		timeout: common.DefaultTimeout,
		result:  tnf.ERROR,
		token:   TokenNotSet,
	}
	for _, o := range options {
		o(as)
	}
	if as.podCheck {
		as.args = []string{"oc", "-n", as.namespace, "get", "pods", as.podname, "-o", "json", "|", "jq", "-r", ".spec"}
	} else {
		as.args = []string{"oc", "-n", as.namespace, "get", "serviceaccounts", as.serviceaccount, "-o", "json"}
	}
	return as
}

// WithNamespace specify the namespace
func WithNamespace(ns string) func(*Automountservice) {
	return func(as *Automountservice) {
		as.namespace = ns
	}
}

//
// WithTimeout specify the timeout of the test
func WithTimeout(t time.Duration) func(*Automountservice) {
	return func(as *Automountservice) {
		as.timeout = t
	}
}

// WithPodname specify the podname to test
func WithPodname(ns string) func(*Automountservice) {
	return func(as *Automountservice) {
		as.podname = ns
		as.podCheck = true
	}
}

// WithServiceAccount specify the serviceaccount to check
func WithServiceAccount(sa string) func(*Automountservice) {
	return func(as *Automountservice) {
		as.serviceaccount = sa
		as.podCheck = false
	}
}

// Args returns the initial execution/send command strings for handler Automountservice.
func (as *Automountservice) Args() []string {
	return as.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (as *Automountservice) GetIdentifier() identifier.Identifier {
	return identifier.AutomountServiceIdentifier
}

// Timeout returns the timeout for the test.
func (as *Automountservice) Timeout() time.Duration {
	return as.timeout
}

// Result returns the test result.
func (as *Automountservice) Result() int {
	return as.result
}

// ReelFirst returns a reel step for handler Automountservice.
func (as *Automountservice) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{saRegex},
		Timeout: as.timeout,
	}
}

// ReelMatch parses the Automountservice output and set the test result on match.
func (as *Automountservice) ReelMatch(_, _, match string) *reel.Step {
	numExpectedMatches := 2
	saMatchIdx := 1
	as.result = tnf.SUCCESS
	re := regexp.MustCompile(saRegex)
	matched := re.FindStringSubmatch(match)
	if len(matched) < numExpectedMatches {
		return nil
	}
	if matched[saMatchIdx] == False {
		as.token = TokenIsFalse
	} else if matched[saMatchIdx] == True {
		as.token = TokenIsTrue
	}
	return nil
}

// ReelTimeout function for Automountservice will be called by the reel FSM when a expect timeout occurs.
func (as *Automountservice) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF function for Automountservice will be called by the reel FSM when a EOF is read.
func (as *Automountservice) ReelEOF() {
}

// Token return the value of automountServiceAccountToken for this test
func (as *Automountservice) Token() int {
	return as.token
}
