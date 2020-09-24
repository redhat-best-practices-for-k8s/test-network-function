package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// whitespaceRegex is a regular expression which matches any whitespace.
var whitespaceRegex = regexp.MustCompile(`(?m)\s`)

var isQuotedRegex = regexp.MustCompile(`(?m)[\"\'].*[\"\']`)

// PrepareString is a utility function used to sanitize a CLI command input.  First, the command arguments are trimmed.
// Next, each trimmed argument is inspected to see if it utilizes white space.  If the argument contains white space, it
// is further checked to ensure that it is contained by single or double quotes.  If it fails this test, then an
// appropriate error is returned.  If all tests pass, the prepared string is created and returned.
func PrepareString(format string, args ...interface{}) (string, error) {
	var preparedStringArgs []interface{}
	for _, arg := range args {
		// only inspect string arguments
		if sArg, ok := arg.(string); ok {
			preparedArg := strings.TrimSpace(sArg)
			whitespaceMatch := whitespaceRegex.FindString(preparedArg)
			if whitespaceMatch != "" {
				quotedMatch := isQuotedRegex.FindString(preparedArg)
				if quotedMatch == "" {
					return "", errors.New(fmt.Sprintf("argument contains non-trimmable whitespace outside of quotes \"%s\"", arg))
				}
			}
			preparedStringArgs = append(preparedStringArgs, preparedArg)
		} else {
			preparedStringArgs = append(preparedStringArgs, arg)
		}
	}
	return fmt.Sprintf(format, preparedStringArgs...), nil
}
