package common

// OcDebugImageID can be set by the test application that uses the test handlers which require the oc debug command to work with a specific image version
var OcDebugImageID string

const (
	ocDebugCmdBase = "oc debug"
)

// GetOcDebugCommand returns the command base for any test handler that uses the oc debug command
func GetOcDebugCommand() string {
	if len(OcDebugImageID) > 0 {
		return ocDebugCmdBase + " --image " + OcDebugImageID
	}
	return ocDebugCmdBase
}
