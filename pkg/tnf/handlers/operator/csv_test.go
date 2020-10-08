package operator_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/operator"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

const (
	testTimeoutDuration = time.Second * 2
	csvName             = "csv-test-v1.0"
	namespace           = "test"
	expectedStatus      = "Succeeded"
)

func TestCsv_Args(t *testing.T) {
	c := operator.NewCsv(csvName, namespace, expectedStatus, testTimeoutDuration)
	args := strings.Split(fmt.Sprintf(operator.CheckCSVCommand, c.Name, c.Namespace), " ")
	fmt.Println(args)
	assert.Equal(t, args, c.Args())
}
func TestCsv_ReelFirst(t *testing.T) {
	c := operator.NewCsv(csvName, namespace, expectedStatus, testTimeoutDuration)
	step := c.ReelFirst()
	assert.Equal(t, "", step.Execute)
	fmt.Println(c.ExpectStatus)
	assert.Equal(t, []string{c.ExpectStatus}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestCsv_ReelEof(t *testing.T) {
	c := operator.NewCsv(csvName, namespace, expectedStatus, testTimeoutDuration)
	// just ensures lack of panic
	c.ReelEOF()
}

func TestCsv_ReelTimeout(t *testing.T) {
	c := operator.NewCsv(csvName, namespace, expectedStatus, testTimeoutDuration)
	step := c.ReelTimeout()
	assert.Nil(t, step)
}
func TestCsv_ReelMatch(t *testing.T) {
	c := operator.NewCsv(csvName, namespace, expectedStatus, testTimeoutDuration)
	step := c.ReelMatch("", "", "Succeeded")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}
func TestNewCsv(t *testing.T) {
	c := operator.NewCsv(csvName, namespace, expectedStatus, testTimeoutDuration)
	assert.Equal(t, tnf.ERROR, c.Result())
	assert.Equal(t, testTimeoutDuration, c.Timeout())
}
