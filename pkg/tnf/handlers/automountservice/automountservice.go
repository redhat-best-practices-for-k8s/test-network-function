// Copyright (C) 2020-2022 Red Hat, Inc.
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

// automountservice is the reel handler struct.
type AutomountService struct {
	result              int
	namespace           string
	isNamespaceSet      bool
	serviceaccount      string
	isServiceAccountSet bool
	podname             string
	isPodnameSet        bool
	timeout             time.Duration
	args                []string
	token               int
}

const (
	allRegex = `(?m).+`
	SaRegex  = `(?m)"automountServiceAccountToken": (.+)`
	False    = "false,"
	True     = "true,"
)

// NewAutomountService returns a new automountservice handler struct.
func NewAutomountService(options ...func(*AutomountService)) *AutomountService {
	as := &AutomountService{
		timeout: common.DefaultTimeout,
		result:  tnf.ERROR,
		token:   TokenNotSet,
	}
	for _, o := range options {
		o(as)
	}
	// to have a valid constructor we need to define
	// namespace and podname Or namespace and serviceaccount
	if !as.isNamespaceSet && (as.isPodnameSet == !as.isServiceAccountSet) {
		return nil
	}

	if as.isPodnameSet {
		as.args = []string{"oc", "-n", as.namespace, "get", "pods", as.podname, "-o", "json", "|", "jq", "-r", ".spec"}
	} else {
		as.args = []string{"oc", "-n", as.namespace, "get", "serviceaccounts", as.serviceaccount, "-o", "json"}
	}
	return as
}

// WithNamespace specify the namespace
func WithNamespace(ns string) func(*AutomountService) {
	return func(as *AutomountService) {
		as.namespace = ns
		as.isNamespaceSet = true
	}
}

// WithTimeout specify the timeout of the test
func WithTimeout(t time.Duration) func(*AutomountService) {
	return func(as *AutomountService) {
		as.timeout = t
	}
}

// WithPodname specify the podname to test
func WithPodname(ns string) func(*AutomountService) {
	return func(as *AutomountService) {
		as.podname = ns
		as.isPodnameSet = true
	}
}

// WithServiceAccount specify the serviceaccount to check
func WithServiceAccount(sa string) func(*AutomountService) {
	return func(as *AutomountService) {
		as.serviceaccount = sa
		as.isServiceAccountSet = true
	}
}

// Args returns the initial execution/send command strings for handler automountservice.
func (as *AutomountService) Args() []string {
	return as.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (as *AutomountService) GetIdentifier() identifier.Identifier {
	return identifier.AutomountServiceIdentifier
}

// Timeout returns the timeout for the test.
func (as *AutomountService) Timeout() time.Duration {
	return as.timeout
}

// Result returns the test result.
func (as *AutomountService) Result() int {
	return as.result
}

// ReelFirst returns a reel step for handler automountservice.
func (as *AutomountService) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{allRegex},
		Timeout: as.timeout,
	}
}

// ReelMatch parses the automountservice output and set the test result on match.
func (as *AutomountService) ReelMatch(_, _, match string) *reel.Step {
	numExpectedMatches := 2
	saMatchIdx := 1
	as.result = tnf.SUCCESS
	re := regexp.MustCompile(SaRegex)
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

// ReelTimeout function for automountservice will be called by the reel FSM when a expect timeout occurs.
func (as *AutomountService) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF function for automountservice will be called by the reel FSM when a EOF is read.
func (as *AutomountService) ReelEOF() {
}

// Token return the value of automountServiceAccountToken for this test
func (as *AutomountService) Token() int {
	return as.token
}
