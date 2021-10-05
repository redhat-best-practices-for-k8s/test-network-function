package common

import "strings"

// OcDebugImageID can be set by the test application that uses the test handlers which require the oc debug command to work with a specific image version
var OcDebugImageID string

const (
	ocCommand    = "oc"
	debugCommand = "debug"
	imageArg     = "--image"
)

// GetOcDebugCommand returns the command base for any test handler that uses the oc debug command
func GetOcDebugCommand() string {
	args := []string{ocCommand, debugCommand}
	if len(OcDebugImageID) > 0 {
		args = append(args, imageArg, OcDebugImageID)
	}
	return strings.Join(args, " ")
}

// GetDebugCommand returns the command base for any test handler that uses the debug command
func GetDebugCommand() string {
	args := []string{debugCommand}
	if len(OcDebugImageID) > 0 {
		args = append(args, imageArg, OcDebugImageID)
	}
	return strings.Join(args, " ")
}
