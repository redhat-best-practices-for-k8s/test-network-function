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

package deploymentsnodes

import (
	"regexp"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	dnRegex      = "(?s).+"
	podNameRegex = "(.+)-[0-9a-z]+-[0-9a-z]+$" // pod name is <deployment>-<pod-template-hash>-<pod-hash>
)

// NodesMap node name to deployments map
type NodesMap map[string]map[string]bool

// DeploymentsNodes holds a mapping of nodes to deployments
type DeploymentsNodes struct {
	nodes   NodesMap // map nodes to deployments
	result  int
	timeout time.Duration
	args    []string
}

// NewDeploymentsNodes creates a new DeploymentsNodes tnf.Test.
func NewDeploymentsNodes(timeout time.Duration, namespace, deploymentName string) *DeploymentsNodes {
	cmd := []string{"oc", "-n", namespace, "get", "pods",
		"-l", "pod-template-hash",
		"-o", "custom-columns=" +
			"NAME:.metadata.name," +
			"NODE:.spec.nodeName",
	}
	if deploymentName != "" {
		cmd = append(cmd, "-l", "app="+deploymentName)
	}
	return &DeploymentsNodes{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    cmd,
		nodes:   NodesMap{},
	}
}

// GetNodes returns nodes to deployments mapping extracted from running the NodesDeployments tnf.Test.
func (dn *DeploymentsNodes) GetNodes() NodesMap {
	return dn.nodes
}

// Args returns the command line args for the test.
func (dn *DeploymentsNodes) Args() []string {
	return dn.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (dn *DeploymentsNodes) GetIdentifier() identifier.Identifier {
	return identifier.DeploymentsNodesIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (dn *DeploymentsNodes) Timeout() time.Duration {
	return dn.timeout
}

// Result returns the test result.
func (dn *DeploymentsNodes) Result() int {
	return dn.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (dn *DeploymentsNodes) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{dnRegex},
		Timeout: dn.timeout,
	}
}

// ReelMatch ensures that list of nodes is not empty and stores the names as []string
func (dn *DeploymentsNodes) ReelMatch(_, _, match string) *reel.Step {
	const (
		numExepctedFields = 2
		podNameIdx        = 0
		nodeNameIdx       = 1
	)
	trimmedMatch := strings.Trim(match, "\n")
	lines := strings.Split(trimmedMatch, "\n")[1:] // First line is the headers/titles line

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != numExepctedFields {
			return nil
		}
		podName := fields[podNameIdx]
		nodeName := fields[nodeNameIdx]
		node, ok := dn.nodes[nodeName]
		if !ok {
			node = map[string]bool{}
			dn.nodes[nodeName] = node
		}
		deploymentName := extractDeployment(podName)
		if deploymentName == "" {
			return nil
		}
		node[deploymentName] = true
	}

	dn.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (dn *DeploymentsNodes) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (dn *DeploymentsNodes) ReelEOF() {
}

func extractDeployment(podName string) string {
	const (
		numExpectedMatches = 2
		deploymentIdx      = 1
	)
	re := regexp.MustCompile(podNameRegex)
	matched := re.FindStringSubmatch(podName)
	if len(matched) != numExpectedMatches {
		return ""
	}
	return matched[deploymentIdx]
}
