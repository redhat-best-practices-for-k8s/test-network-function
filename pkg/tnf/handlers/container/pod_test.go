package container_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container/testcases"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

const (
	testTimeoutDuration = time.Second * 2
	name                = "HOST_NETWORK_CHECK"
	namespace           = "test"
	command             = "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostNetwork'"
	actionAllow         = "allow"
	actionDeny          = "deny"
)

var (
	stringExpectedStatus = []string{"NULL_FALSE"}
	sliceExpectedStatus  = []string{"NET_ADMIN", "SYS_TIME"}
)

func TestPod_Args(t *testing.T) {
	c := container.NewPod(command, name, namespace, stringExpectedStatus, actionAllow, "string", testTimeoutDuration)
	args := strings.Split(fmt.Sprintf(command, c.Name, c.Namespace), " ")
	assert.Equal(t, args, c.Args())
}

func TestPod_ReelFirst(t *testing.T) {
	c := container.NewPod(command, name, namespace, stringExpectedStatus, actionAllow, "string", testTimeoutDuration)
	step := c.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Equal(t, []string{testcases.GetOutRegExp("ALLOW_ALL")}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestPod_ReelEof(t *testing.T) {
	c := container.NewPod(command, name, namespace, stringExpectedStatus, "string", actionAllow, testTimeoutDuration)
	// just ensures lack of panic
	c.ReelEOF()
}

func TestPod_ReelTimeout(t *testing.T) {
	c := container.NewPod(command, name, namespace, stringExpectedStatus, "string", actionAllow, testTimeoutDuration)
	step := c.ReelTimeout()
	assert.Nil(t, step)
}
func TestPodTest_ReelMatchString(t *testing.T) {
	c := container.NewPod(command, name, namespace, stringExpectedStatus, "string", actionAllow, testTimeoutDuration)
	step := c.ReelMatch("", "", "null")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatchStringNoFound(t *testing.T) {
	c := container.NewPod(command, name, namespace, stringExpectedStatus, "string", actionAllow, testTimeoutDuration)
	step := c.ReelMatch("", "", "not_null")
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestPodTest_ReelMatchArray_Allow_Match(t *testing.T) {
	c := container.NewPod(command, name, namespace, sliceExpectedStatus, "array", actionAllow, testTimeoutDuration)
	step := c.ReelMatch("", "", `["NET_ADMIN", "SYS_TIME"]`)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatchArray_Allow_NoMatch(t *testing.T) {
	c := container.NewPod(command, name, namespace, sliceExpectedStatus, "array", actionAllow, testTimeoutDuration)
	step := c.ReelMatch("", "", `["NET_ADMIN", "NO_SYS_TIME"]`)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}
func TestPodTest_ReelMatchArray_Deny_Match(t *testing.T) {
	c := container.NewPod(command, name, namespace, sliceExpectedStatus, "array", actionDeny, testTimeoutDuration)
	step := c.ReelMatch("", "", `["NET_ADMIN", "SYS_TIME"]`)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}
func TestPodTest_ReelMatchArray_Deny_NotMatch(t *testing.T) {
	c := container.NewPod(command, name, namespace, sliceExpectedStatus, "array", actionDeny, testTimeoutDuration)
	step := c.ReelMatch("", "", `["NOT_NET_ADMIN", "NOT_SYS_TIME"]`)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}
func TestNewPod(t *testing.T) {
	c := container.NewPod(command, name, namespace, stringExpectedStatus, "string", actionAllow, testTimeoutDuration)
	assert.Equal(t, tnf.ERROR, c.Result())
	assert.Equal(t, testTimeoutDuration, c.Timeout())
}
