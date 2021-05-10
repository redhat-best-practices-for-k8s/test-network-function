package cr_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/example-cnf/cr"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"testing"
)

func TestTraffic_Args(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	assert.Equal(t, []string{"oc", "describe", "-n", "ns", "crtype", "crname", "|", "grep", "TestPassed"}, traffic.Args())
}

func TestTraffic_GetIdentifier(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	identifier := traffic.GetIdentifier()
	assert.NotNil(t, identifier)
	assert.Equal(t, "http://test-network-function.com/test-network-function/cr/create", identifier.URL)
}

func TestTraffic_Timeout(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	assert.Equal(t, testTimeout, traffic.Timeout())
}

func TestTraffic_ReelFirst(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	step := traffic.ReelFirst()
	assert.NotNil(t, step)
}

func TestTraffic_ReelMatch(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	assert.Equal(t, tnf.ERROR, traffic.Result())
	step := traffic.ReelMatch("", "", "")
	assert.Nil(t, step)
}

func TestTraffic_ReelTimeout(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	step := traffic.ReelTimeout()
	assert.Nil(t, step)
}

func TestTraffic_ReelEOF(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	// just ensure no panic
	traffic.ReelEOF()
}

func TestNewTraffic(t *testing.T) {
	traffic := cr.NewTraffic("ns", "crtype", "crname", testTimeout)
	assert.NotNil(t, traffic)
}
