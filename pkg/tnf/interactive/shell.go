package interactive

import (
	expect "github.com/google/goexpect"
	"os"
	"time"
)

const (
	shellEnvironmentVariableKey = "SHELL"
)

// Creates an interactive shell subprocess based on the value of $SHELL, spawning the appropriate underlying PTY.
func SpawnShell(spawner *Spawner, timeout time.Duration, opts ...expect.Option) (*Context, error) {
	shellEnv := os.Getenv(shellEnvironmentVariableKey)
	var args []string
	return (*spawner).Spawn(shellEnv, args, timeout, opts...)
}
