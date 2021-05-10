package cr_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/example-cnf/cr"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"testing"
	"time"
)

var (
	testTimeout = time.Second * 1
)

func TestCreate_Args(t *testing.T) {
	create := cr.NewCreate("ex.yaml", testTimeout)
	assert.Equal(t, []string{"oc", "create", "-f", "ex.yaml"}, create.Args())
}

func TestCreate_GetIdentifier(t *testing.T) {
	create := cr.NewCreate("ex.yaml", testTimeout)
	identifier := create.GetIdentifier()
	assert.NotNil(t, identifier)
	assert.Equal(t, "http://test-network-function.com/test-network-function/cr/create", identifier.URL)
}

func TestCreate_Timeout(t *testing.T) {
	create := cr.NewCreate("ex.yaml", testTimeout)
	assert.Equal(t, testTimeout, create.Timeout())
}

func TestCreate_ReelFirst(t *testing.T) {
	create := cr.NewCreate("ex.yaml", testTimeout)
	assert.NotNil(t, create.ReelFirst())
}

func TestCreate_ReelMatch(t *testing.T) {
	create := cr.NewCreate("ex.yaml", testTimeout)
	assert.Equal(t, tnf.ERROR, create.Result())
	step := create.ReelMatch("", "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, create.Result())
}

func TestCreate_ReelTimeout(t *testing.T) {
	create := cr.NewCreate("ex.yaml", testTimeout)
	step := create.ReelTimeout()
	assert.Nil(t, step)
}

func TestCreate_ReelEOF(t *testing.T) {
	create := cr.NewCreate("ex.yaml", testTimeout)
	// just ensure no panics
	create.ReelEOF()
}

func TestNewCreate(t *testing.T) {
	create := cr.NewCreate("", testTimeout)
	assert.NotNil(t, create)
}
