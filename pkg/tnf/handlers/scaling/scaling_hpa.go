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

package scaling

import (
	"fmt"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	hpaOcCommand = "oc patch hpa %s  -p '{\"spec\":{\"minReplicas\": %d, \"maxReplicas\": %d}}' -n %s"
	hpaRegex     = "horizontalpodautoscaler.autoscaling/%s patched"
)

// Scaling holds the Scaling handler parameters.
type HpAScaling struct {
	result  int
	timeout time.Duration
	args    []string
	regex   string
}

// NewScaling creates a new Scaling handler.
func HpaNewScaling(timeout time.Duration, namespace, hpaName string, min, max int) *HpAScaling {
	command := fmt.Sprintf(hpaOcCommand, hpaName, min, max, namespace)
	return &HpAScaling{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    strings.Fields(command),
		regex:   fmt.Sprintf(hpaRegex, hpaName),
	}
}

// Args returns the command line args for the test.
func (hpascaling *HpAScaling) Args() []string {
	return hpascaling.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (hpascaling *HpAScaling) GetIdentifier() identifier.Identifier {
	return identifier.ScalingIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (hpascaling *HpAScaling) Timeout() time.Duration {
	return hpascaling.timeout
}

// Result returns the test result.
func (hpascaling *HpAScaling) Result() int {
	return hpascaling.result
}

// ReelFirst returns a step which expects the scale command output within the test timeout.
func (hpascaling *HpAScaling) ReelFirst() *reel.Step {
	return &reel.Step{
		Execute: "",
		Expect:  []string{hpascaling.regex},
		Timeout: hpascaling.timeout,
	}
}

// ReelMatch does nothing, just set the test result as success.
func (hpascaling *HpAScaling) ReelMatch(_, _, match string) *reel.Step {
	hpascaling.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (hpascaling *HpAScaling) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (hpascaling *HpAScaling) ReelEOF() {
}
