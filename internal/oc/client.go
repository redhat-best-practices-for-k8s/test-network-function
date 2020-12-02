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

package oc

import "github.com/redhat-nfvpe/test-network-function/internal/subprocess"

const (
	containerCommandSeparator = "--"
	ocExecCommand             = "exec"
	ocExecContainerArg        = "-c"
	ocCommand                 = "oc"
	ocNamespaceArg            = "-n"
)

// InvokeOCCommand is a lightweight wrapper client around oc client.
func InvokeOCCommand(pod, container, namespace string, command []string) (string, error) {
	invokeCommandArgs := []string{ocExecCommand, pod, ocExecContainerArg, container}
	if namespace != "" {
		invokeCommandArgs = append(invokeCommandArgs, ocNamespaceArg, namespace)
	}
	invokeCommandArgs = append(invokeCommandArgs, containerCommandSeparator)
	invokeCommandArgs = append(invokeCommandArgs, command...)
	return subprocess.InvokeCommand(ocCommand, invokeCommandArgs)
}
