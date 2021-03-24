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

package clusterrolebinding

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	crbRegex = "(?s).+"
)

type ClusterRoleBinding struct {
	clusterRoleBindings []string // Output variable that stores the 'bad' ClusterRoleBindings
	result              int
	timeout             time.Duration
	args                []string
}

func NewClusterRoleBinding(timeout time.Duration, serviceAccountName, podNamespace string) *ClusterRoleBinding {
	serviceAccountSubString := "name:" + serviceAccountName + " namespace:" + podNamespace
	return &ClusterRoleBinding{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"oc get clusterrolebindings -o custom-columns='NAME:metadata.name,SERVICE_ACCOUNTS:subjects[?(@.kind==\"ServiceAccount\")]' | grep -E '" +
				serviceAccountSubString +
				"|SERVICE_ACCOUNTS'"},
	}
}

func (crb *ClusterRoleBinding) GetClusterRoleBindings() []string {
	return crb.clusterRoleBindings
}

// Args returns the command line args for the test.
func (crb *ClusterRoleBinding) Args() []string {
	return crb.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (crb *ClusterRoleBinding) GetIdentifier() identifier.Identifier {
	return identifier.ClusterRoleBindingIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (crb *ClusterRoleBinding) Timeout() time.Duration {
	return crb.timeout
}

// Result returns the test result.
func (crb *ClusterRoleBinding) Result() int {
	return crb.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (crb *ClusterRoleBinding) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{crbRegex},
		Timeout: crb.timeout,
	}
}

func (crb *ClusterRoleBinding) ReelMatch(_, _, match string) *reel.Step {
	const (
		nameIdx = 0
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

		crb.clusterRoleBindings = append(crb.clusterRoleBindings, fields[nameIdx])
	}

	if len(crb.clusterRoleBindings) == 0 {
		crb.result = tnf.SUCCESS
	} else {
		crb.result = tnf.FAILURE
	}

	return nil
}

func (crb *ClusterRoleBinding) ReelTimeout() *reel.Step {
	return nil
}

func (crb *ClusterRoleBinding) ReelEOF() {
}
