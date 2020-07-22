package subprocess

import (
	"os/exec"
)

// Lightweight wrapper around command subprocess which waits for the underlying process to complete prior to returning.
func InvokeCommand(executable string, args []string) (string, error) {
	cmd := exec.Command(executable, args...)
	stdout, err := cmd.Output()
	return string(stdout), err
}
