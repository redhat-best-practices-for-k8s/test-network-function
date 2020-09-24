package utils_test

import (
	"errors"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

type prepareStringTest struct {
	format         string
	args           []interface{}
	expectedError  error
	expectedOutput string
}

var prepareStringTestCases = map[string]prepareStringTest{
	"no_arguments": {
		format:         "ls",
		args:           iface([]string{}),
		expectedError:  nil,
		expectedOutput: "ls",
	},
	"acceptable_arguments": {
		format:         "ls %s",
		args:           iface([]string{"/"}),
		expectedError:  nil,
		expectedOutput: "ls /",
	},
	"acceptable_multiple_arguments": {
		format:         "ssh %s@%s",
		args:           iface([]string{"user", "host"}),
		expectedError:  nil,
		expectedOutput: "ssh user@host",
	},
	"leading_and_trailing_whitespace_should_be_fine": {
		format:         "ssh %s@%s",
		args:           iface([]string{"    user", "host     "}),
		expectedError:  nil,
		expectedOutput: "ssh user@host",
	},
	"mixed_argument_types": {
		format:         "ping -c %d %s",
		args:           nil,
		expectedError:  nil,
		expectedOutput: "ping -c 1 host",
	},
	"empty_string": {
		format: "%s",
		args: iface([]string{" "}),
		expectedError: nil,
		expectedOutput: "",
	},
	"double_quoted_argument": {
		format: "echo %s",
		args: iface([]string{"\"   I can do this fine   \""}),
		expectedError: nil,
		expectedOutput: "echo \"   I can do this fine   \"",
	},
	"single_quoted_argument_unescaped": {
		format: "echo %s",
		args: iface([]string{"'   I can do this fine   '"}),
		expectedError: nil,
		expectedOutput: "echo '   I can do this fine   '",
	},
	// Negative Tests.
	"negative_test_whitespace_arg": {
		format:         "ssh %s@%s",
		args:           iface([]string{"user 1", "host"}),
		expectedError:  errors.New("argument contains non-trimmable whitespace outside of quotes \"user 1\""),
		expectedOutput: "",
	},
	"negative_test_non_terminated_double_quote": {
		format: "echo %s",
		args: iface([]string{"\"   hi "}),
		expectedError: errors.New("argument contains non-trimmable whitespace outside of quotes \"\"   hi \""),
		expectedOutput: "",
	},
	"negative_test_non_started_double_quote": {
		format: "echo %s",
		args: iface([]string{"   hi \""}),
		expectedError: errors.New("argument contains non-trimmable whitespace outside of quotes \"   hi \"\""),
		expectedOutput: "",
	},
	"negative_test_non_terminated_single_quote": {
		format: "echo %s",
		args: iface([]string{"'   hi "}),
		expectedError: errors.New("argument contains non-trimmable whitespace outside of quotes \"'   hi \""),
		expectedOutput: "",
	},
	"negative_test_non_started_single_quote": {
		format: "echo %s",
		args: iface([]string{"   hi '"}),
		expectedError: errors.New("argument contains non-trimmable whitespace outside of quotes \"   hi '\""),
		expectedOutput: "",
	},
}

func TestPrepareString(t *testing.T) {
	var mixedArguments []interface{}
	mixedArguments = append(mixedArguments, 1)
	mixedArguments = append(mixedArguments, "host")

	for testName, testCase := range prepareStringTestCases {
		var inputArgs []interface{}
		inputArgs = testCase.args
		if testName == "mixed_argument_types" {
			inputArgs = mixedArguments
		}
		actualOutput, actualErr := utils.PrepareString(testCase.format, inputArgs...)
		assert.Equal(t, testCase.expectedError, actualErr)
		assert.Equal(t, testCase.expectedOutput, actualOutput)
	}
}

// iface is a utility function to convert a string array into an interface array for compatibility with fmt APIs.
func iface(list []string) []interface{} {
	vals := make([]interface{}, len(list))
	for i, v := range list {
		vals[i] = v
	}
	return vals
}
