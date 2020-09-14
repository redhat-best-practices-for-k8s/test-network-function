package hostname_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/hostname"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testTimeoutDuration = time.Second * 2
)

func TestHostname_Args(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	assert.Equal(t, []string{"hostname"}, h.Args())
}

func TestHostname_ReelFirst(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	step := h.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Equal(t, []string{hostname.SuccessfulOutputRegex}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestHostname_ReelEof(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	// just ensures lack of panic
	h.ReelEOF()
}

func TestHostname_ReelTimeout(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	step := h.ReelTimeout()
	assert.Nil(t, step)
}

// Also tests GetHostname() and Result()
func TestHostname_ReelMatch(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	matchHostname := "testHostname"
	step := h.ReelMatch("", "", matchHostname)
	assert.Nil(t, step)
	assert.Equal(t, matchHostname, h.GetHostname())
	assert.Equal(t, tnf.SUCCESS, h.Result())
}

func TestNewHostname(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	assert.Equal(t, tnf.ERROR, h.Result())
	assert.Equal(t, testTimeoutDuration, h.Timeout())
}
