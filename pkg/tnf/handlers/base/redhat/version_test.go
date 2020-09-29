package redhat_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

var testTimeoutDuration = time.Second * 2

// TestNewRelease also tests Args, Timeout and Result
func TestNewRelease(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)
	assert.NotNil(t, r)
	assert.Equal(t, strings.Split(redhat.ReleaseCommand, " "), r.Args())
	assert.Equal(t, testTimeoutDuration, r.Timeout())
	assert.Equal(t, tnf.ERROR, r.Result())
}

func TestRelease_ReelFirst(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)
	step := r.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Contains(t, step.Expect, redhat.VersionRegex)
	assert.Contains(t, step.Expect, redhat.NotRedHatBasedRegex)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestRelease_ReelMatch(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)

	// Positive test.
	step := r.ReelMatch(redhat.VersionRegex, "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, r.Result())

	r = redhat.NewRelease(testTimeoutDuration)

	// Negative test.
	step = r.ReelMatch(redhat.NotRedHatBasedRegex, "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, r.Result())

	// Error case.  Note, this shouldn't ever happen based on the FSM, but it is better to be defensive.
	step = r.ReelMatch("unknown regex", "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, r.Result())
}

func TestRelease_ReelTimeout(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)
	step := r.ReelTimeout()
	assert.Nil(t, step)
}

func TestRelease_ReelEof(t *testing.T) {
	// just ensures no panics
	r := redhat.NewRelease(testTimeoutDuration)
	r.ReelEOF()
}
