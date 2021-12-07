package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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
// its output wihout any other check.
func ExecuteCommand(command string, timeout time.Duration, context *interactive.Context, failureCallbackFun func()) string {
	log.Debugf("Executing command: %s", command)

	values := make(map[string]interface{})
	// Escapes the double quote and new line chars to make a valid json string for the command to be executed by the handler.
	var err error
	values["COMMAND"], err = escapeToJSONstringFormat(command)
	gomega.Expect(err).To(gomega.BeNil())
	values["TIMEOUT"] = timeout.Nanoseconds()

	log.Debugf("Command handler's COMMAND string value: %s", values["COMMAND"])

	tester, handlers := NewGenericTestAndValidate(commandHandlerFilePath, handlerJSONSchemaFilePath, values)
	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	test.RunAndValidateWithFailureCallback(failureCallbackFun)

	genericTest := (*tester).(*generic.Generic)
	gomega.Expect(genericTest).ToNot(gomega.BeNil())

	matches := genericTest.Matches
	gomega.Expect(len(matches)).To(gomega.Equal(1))
	match := genericTest.GetMatches()[0]
	return match.Match
}

// NewGenericTestAndValidate creates a generic handler from the json template with the var map and validate the outcome
func NewGenericTestAndValidate(templateFile, schemaPath string, values map[string]interface{}) (*tnf.Tester, []reel.Handler) {
	tester, handlers, result, err := generic.NewGenericFromMap(templateFile, handlerJSONSchemaFilePath, values)

	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	return tester, handlers
}

func RunCommandInContainerNameSpace(nodeName string, nodeOc *interactive.Oc, containerID, command string, timeout time.Duration, isNonOcp bool) string {
	containrPID := GetContainerPID(nodeName, nodeOc, containerID, isNonOcp)
	nodeCommand := "nsenter -t " + containrPID + " -n " + command
	return RunCommandInNode(nodeName, nodeOc, nodeCommand, timeout)
}

func GetContainerPID(nodeName string, nodeOc *interactive.Oc, containerID string, isNonOcp bool) string {
	command := ""
	if isNonOcp {
		command = "chroot /host docker inspect -f '{{.State.Pid}}' " + containerID + " 2>/dev/null"
	} else {
		command = "chroot /host crictl inspect --output go-template --template '{{.info.pid}}' " + containerID + " 2>/dev/null"
	}
	return RunCommandInNode(nodeName, nodeOc, command, timeoutPid)
}
func RunCommandInNode(nodeName string, nodeOc *interactive.Oc, command string, timeout time.Duration) string {
	context := nodeOc
	tester := nodedebug.NewNodeDebug(timeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	return tester.Raw
}

func AddNsenterPrefix(containerPID string) string {
	return "chroot /host nsenter -t " + containerPID + " -n "
}
