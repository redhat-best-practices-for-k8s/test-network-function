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

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodedebug"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

var (
	// pathRelativeToRoot is used to calculate relative filepaths to the tnf folder.
	pathRelativeToRoot = path.Join("..")
	// commandHandlerFilePath is the file location of the command handler.
	commandHandlerFilePath = path.Join(pathRelativeToRoot, "pkg", "tnf", "handlers", "command", "command.json")
	// handlerJSONSchemaFilePath is the file location of the json handlers generic schema.
	handlerJSONSchemaFilePath = path.Join(pathRelativeToRoot, "schemas", "generic-test.schema.json")
)

const (
	timeoutPid = 5 * time.Second
)

// ArgListToMap takes a list of strings of the form "key=value" and translate it into a map
// of the form {key: value}
func ArgListToMap(lst []string) map[string]string {
	retval := make(map[string]string)
	for _, arg := range lst {
		splitArgs := strings.Split(arg, "=")
		if len(splitArgs) == 1 {
			retval[splitArgs[0]] = ""
		} else {
			retval[splitArgs[0]] = splitArgs[1]
		}
	}
	return retval
}

// FilterArray takes a list and a predicate and returns a list of all elements for whom the predicate returns true
func FilterArray(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func CheckFileExists(filePath, name string) {
	fullPath, _ := filepath.Abs(filePath)
	if _, err := os.Stat(fullPath); err == nil {
		log.Infof("Path to %s file found and valid: %s ", name, fullPath)
	} else if errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Path to %s file not found: %s , Exiting", name, fullPath)
	} else {
		log.Fatalf("Path to %s file not valid: %s , err=%s, exiting", name, fullPath, err)
	}
}

func escapeToJSONstringFormat(line string) (string, error) {
	// Newlines need manual escaping.
	line = strings.ReplaceAll(line, "\n", "\\n")
	marshalled, err := json.Marshal(line)
	if err != nil {
		return "", err
	}
	s := string(marshalled)
	// Remove double quotes and return marshalled string.
	return s[1 : len(s)-1], nil
}

// ExecuteCommand uses the generic command handler to execute an arbitrary interactive command, returning
// its output wihout any filtering/matching if the command is successfully executed
func ExecuteCommand(command string, timeout time.Duration, context *interactive.Context) (string, error) {
	tester, test := newGenericCommandTester(command, timeout, context)
	result, err := test.Run()
	if result == tnf.SUCCESS && err == nil {
		genericTest := (*tester).(*generic.Generic)
		if genericTest != nil {
			matches := genericTest.Matches
			if len(matches) == 1 {
				return genericTest.GetMatches()[0].Match, nil
			}
		}
	}
	return "", err
}

// ExecuteCommandAndValidate uses the generic command handler to execute an arbitrary interactive command, returning
// its output wihout any filtering/matching
var ExecuteCommandAndValidate = func(command string, timeout time.Duration, context *interactive.Context, failureCallbackFun func()) string {
	tester, test := newGenericCommandTester(command, timeout, context)
	test.RunAndValidateWithFailureCallback(failureCallbackFun)
	genericTest := (*tester).(*generic.Generic)
	gomega.Expect(genericTest).ToNot(gomega.BeNil())

	matches := genericTest.Matches
	gomega.Expect(len(matches)).To(gomega.Equal(1))
	match := genericTest.GetMatches()[0]
	return match.Match
}

func newGenericCommandTester(command string, timeout time.Duration, context *interactive.Context) (*tnf.Tester, *tnf.Test) {
	log.Debugf("Executing command: %s", command)

	values := make(map[string]interface{})
	// Escapes the double quote and new line chars to make a valid json string for the command to be executed by the handler.
	var err error
	values["COMMAND"], err = escapeToJSONstringFormat(command)
	gomega.Expect(err).To(gomega.BeNil())
	values["TIMEOUT"] = timeout.Nanoseconds()

	log.Debugf("Command handler's COMMAND string value: %s", values["COMMAND"])

	tester, handlers := NewGenericTesterAndValidate(commandHandlerFilePath, handlerJSONSchemaFilePath, values)
	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())
	return tester, test
}

// NewGenericTesterAndValidate creates a generic handler from the json template with the var map and validate the outcome
func NewGenericTesterAndValidate(templateFile, schemaPath string, values map[string]interface{}) (*tnf.Tester, []reel.Handler) {
	tester, handlers, result, err := generic.NewGenericFromMap(templateFile, handlerJSONSchemaFilePath, values)

	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	return tester, handlers
}

// GetContainerPID gets the container PID from a kubernetes node, Oc and container PID
func GetContainerPID(nodeName string, nodeOc *interactive.Oc, containerID, runtime string) string {
	command := ""
	switch runtime {
	case "docker": //nolint:goconst // used only once
		command = "chroot /host docker inspect -f '{{.State.Pid}}' " + containerID + " 2>/dev/null"
	case "docker-pullable": //nolint:goconst // used only once
		command = "chroot /host docker inspect -f '{{.State.Pid}}' " + containerID + " 2>/dev/null"
	case "cri-o", "containerd": //nolint:goconst // used only once
		command = "chroot /host crictl inspect --output go-template --template '{{.info.pid}}' " + containerID + " 2>/dev/null"
	default:
		ginkgo.Skip(fmt.Sprintf("Container runtime %s not supported yet for this test, skipping", runtime))
	}
	return RunCommandInNode(nodeName, nodeOc, command, timeoutPid)
}

func GetModulesFromNode(nodeName string, nodeOc *interactive.Oc) []string {
	// Get the 1st column list of the modules running on the node.
	// Split on the return/newline and get the list of the modules back.
	//nolint:goconst // used only once
	command := `chroot /host lsmod | awk '{ print $1 }' | grep -v Module`
	output := RunCommandInNode(nodeName, nodeOc, command, timeoutPid)
	output = strings.ReplaceAll(output, "\t", "")
	return strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
}

func ModuleInTree(nodeName, moduleName string, nodeOc *interactive.Oc) bool {
	command := `chroot /host modinfo ` + moduleName + ` | awk '{ print $1 }'`
	cmdOutput := RunCommandInNode(nodeName, nodeOc, command, timeoutPid)
	outputSlice := strings.Split(strings.ReplaceAll(cmdOutput, "\r\n", "\n"), "\n")
	// The output, if found, should look something like 'intree:   Y'.
	// As long as we look for 'intree:' being contained in the string we should be good to go.
	found := false
	if StringInSlice(outputSlice, `intree:`, false) {
		found = true
	}
	return found
}

// RunCommandInNode runs a command on a remote kubernetes node
// takes the node name, node oc and command
// returns the command raw output
var RunCommandInNode = func(nodeName string, nodeOc *interactive.Oc, command string, timeout time.Duration) string {
	context := nodeOc
	tester := nodedebug.NewNodeDebug(timeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	return tester.Raw
}

// AddNsenterPrefix adds the nsenter command prefix to run inside a container namespace
func AddNsenterPrefix(containerPID string) string {
	return "nsenter -t " + containerPID + " -n "
}

// StringInSlice checks a slice for a given string.
func StringInSlice(s []string, str string, contains bool) bool {
	for _, v := range s {
		if !contains {
			if strings.TrimSpace(v) == str {
				return true
			}
		} else {
			if strings.Contains(strings.TrimSpace(v), str) {
				return true
			}
		}
	}
	return false
}
