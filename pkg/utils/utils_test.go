package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testString1        = "{{\"Quoted line with new line\n char and also some others \b chars not commonly used like \f, \t, \\ and \r.\"}}"
	testEscapedString1 = `{{\"Quoted line with new line\n char and also some others \u0008 chars not commonly used like \u000c, \t, \\ and \r.\"}}`
)

func TestEscapeToJSONstringFormat(t *testing.T) {
	escapedString, err := escapeToJSONstringFormat(testString1)
	assert.Nil(t, err)
	assert.Equal(t, testEscapedString1, escapedString)
}
