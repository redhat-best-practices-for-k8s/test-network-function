package common

import "strings"

const (
	ocCommand    = "oc"
	debugCommand = "debug"
)

// GetOcDebugCommand returns the command base for any test handler that uses the oc debug command
func GetOcDebugCommand() string {
	args := []string{ocCommand, debugCommand}
	return strings.Join(args, " ")
}

// GetDebugCommand returns the command base for any test handler that uses the debug command
func GetDebugCommand() string {
	args := []string{debugCommand}
	return strings.Join(args, " ")
}
