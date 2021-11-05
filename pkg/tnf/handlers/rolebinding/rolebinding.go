// Copyright (C) 2021 Red Hat, Inc.
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

package rolebinding

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	rbRegex = "(?s).+"
)

// RoleBinding holds information derived from running "oc get rolebindings" on the command line.
type RoleBinding struct {
	podNamespace string
	roleBindings []string // Output variable that stores the 'bad' RoleBindings
	result       int
	timeout      time.Duration
	args         []string
}

// NewRoleBinding creates a new RoleBinding tnf.Test.
func NewRoleBinding(timeout time.Duration, serviceAccountName, podNamespace string) *RoleBinding {
	serviceAccountSubString := "name:" + serviceAccountName + " namespace:" + podNamespace
	return &RoleBinding{
		podNamespace: podNamespace,
		timeout:      timeout,
		result:       tnf.ERROR,
		args: []string{
			"oc get rolebindings --all-namespaces -o custom-columns='NAMESPACE:metadata.namespace,NAME:metadata.name,SERVICE_ACCOUNTS:subjects[?(@.kind==\"ServiceAccount\")]' | grep -E '" +
				serviceAccountSubString +
				"|SERVICE_ACCOUNTS'"},
	}
}

// GetRoleBindings returns any role bindings extracted from running the RoleBindings tnf.Test.
func (rb *RoleBinding) GetRoleBindings() []string {
	return rb.roleBindings
}

// Args returns the command line args for the test.
func (rb *RoleBinding) Args() []string {
	return rb.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (rb *RoleBinding) GetIdentifier() identifier.Identifier {
	return identifier.RoleBindingIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (rb *RoleBinding) Timeout() time.Duration {
	return rb.timeout
}

// Result returns the test result.
func (rb *RoleBinding) Result() int {
	return rb.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (rb *RoleBinding) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{rbRegex},
		Timeout: rb.timeout,
	}
}

// ReelMatch ensures that there are no ServiceAccount RoleBindings for a given OpenShift Pod namespace.
func (rb *RoleBinding) ReelMatch(_, _, match string) *reel.Step {
	const (
		nsIdx   = 0
		nameIdx = 1
	)

	lines := strings.Split(match, "\n")[1:] // First line is the headers/titles line

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		// RoleBinding in pod namespace is allowed and not saved
		if fields[nsIdx] == rb.podNamespace {
			continue
		}

		// RoleBinding in different namespace is saved for reporting failures
		rb.roleBindings = append(rb.roleBindings, fields[nsIdx]+":"+fields[nameIdx])
	}

	if len(rb.roleBindings) == 0 {
		rb.result = tnf.SUCCESS
	} else {
		rb.result = tnf.FAILURE
	}

	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (rb *RoleBinding) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (rb *RoleBinding) ReelEOF() {
}
