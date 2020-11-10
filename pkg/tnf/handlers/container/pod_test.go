package container_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/testcases"
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
)

var (
	stringExpectedStatus             = []string{string(testcases.NullFalse)}
	sliceExpectedStatus              = []string{"NET_ADMIN", "SYS_TIME"}
	resultSliceExpectedStatus        = `["NET_ADMIN", "SYS_TIME"]`
	resultSliceExpectedStatusInvalid = `["NO_NET_ADMIN", "NO_SYS_TIME"]`
	IsNull                           = "null"
	IsNotNull                        = "not_null"
	args                             = strings.Split(fmt.Sprintf(command, name, namespace), " ")
)

func TestPod_Args(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, args, c.Args())
}

func TestPod_ReelFirst(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Equal(t, []string{testcases.GetOutRegExp(testcases.AllowAll)}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestPod_ReelEof(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	// just ensures lack of panic
	c.ReelEOF()
}

func TestPod_ReelTimeout(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelTimeout()
	assert.Nil(t, step)
}

func TestPodTest_ReelMatch_String(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNull)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_Facts(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNull)
	assert.Nil(t, step)
	assert.NotNil(t, c.Facts())
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatch_String_NotFound(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNotNull)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestPodTest_ReelMatch_Array_Allow_Deny_ISNULL(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNull)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatchArray_Allow_Match(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatus)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatch_Array_Allow_NoMatch(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatusInvalid)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestPodTest_ReelMatch_Array_Deny_Match(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Deny, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatus)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestPodTest_ReelMatch_Array_Deny_NotMatch(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Deny, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatusInvalid)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestNewPod(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, tnf.ERROR, c.Result())
	assert.Equal(t, testTimeoutDuration, c.Timeout())
}
