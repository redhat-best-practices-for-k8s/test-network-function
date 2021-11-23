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

package imagepullpolicy

import (
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

// Imagepullpolicy is the reel handler struct.
type Imagepullpolicy struct {
	result  int
	timeout time.Duration
	args    []string
	regex   string
	// adding special parameters
}

const (
	ocCommand = "oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].imagePullPolicy'"
	regex     = "IfNotPresent"
)

// NewImagepullpolicy returns a new Imagepullpolicy handler struct.
// TODO: Add needed parameters to this function and initialize the handler properly.
func NewImagepullpolicy(timeout time.Duration, namespace, podName string, index int) *Imagepullpolicy {
	command := fmt.Sprintf(ocCommand, podName, namespace, index)
	return &Imagepullpolicy{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    strings.Fields(command),
		regex:   regex,
	}
}

// Args returns the initial execution/send command strings for handler Imagepullpolicy.
func (imagepullpolicy *Imagepullpolicy) Args() []string {
	return imagepullpolicy.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (imagepullpolicy *Imagepullpolicy) GetIdentifier() identifier.Identifier {
	// Return the Imagepullpolicy handler identifier.
	return identifier.ImagePullPolicyIdentifier
}

// Timeout returns the timeout for the test.
func (imagepullpolicy *Imagepullpolicy) Timeout() time.Duration {
	return imagepullpolicy.timeout
}

// Result returns the test result.
func (imagepullpolicy *Imagepullpolicy) Result() int {
	return imagepullpolicy.result
}

// ReelFirst returns a reel step for handler Imagepullpolicy.
func (imagepullpolicy *Imagepullpolicy) ReelFirst() *reel.Step {
	return &reel.Step{
		Execute: "",
		Expect:  []string{imagepullpolicy.regex}, // TODO : pass the list of possible regex in here
		Timeout: imagepullpolicy.timeout,
	}
}

// ReelMatch parses the Imagepullpolicy output and set the test result on match.
func (imagepullpolicy *Imagepullpolicy) ReelMatch(_, _, match string) *reel.Step {
	// TODO : add the matching logic here and return an appropriate tnf result.
	if match == imagepullpolicy.regex {
		ginkgo.By(fmt.Sprintf("%s equal to %s ", match, imagepullpolicy.regex))
		imagepullpolicy.result = tnf.SUCCESS
	} else {
		imagepullpolicy.result = tnf.FAILURE
		ginkgo.By(fmt.Sprintf("%s is not equal to %s ", match, imagepullpolicy.regex))
	}
	return nil
}

// ReelTimeout function for Imagepullpolicy will be called by the reel FSM when a expect timeout occurs.
func (imagepullpolicy *Imagepullpolicy) ReelTimeout() *reel.Step {
	// TODO : Add code here in case a timeout reaction is needed.
	return nil
}

// ReelEOF function for Imagepullpolicy will be called by the reel FSM when a EOF is read.
func (imagepullpolicy *Imagepullpolicy) ReelEOF() {
	// TODO : Add code here in case a EOF reaction is needed.
}
