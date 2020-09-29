package interactive

import (
	"fmt"
	expect "github.com/google/goexpect"
	"time"
)

const (
	sshCommand   = "ssh"
	sshSeparator = "@"
)

// SpawnSSH spawns an SSH session to a generic linux host using ssh provided by openssh-clients.  Takes care of
// establishing the pseudo-terminal (PTY) through expect.SpawnGeneric().
// TODO: This method currently relies upon passwordless SSH setup beforehand.  Handle all types of auth.
func SpawnSSH(spawner *Spawner, user, host string, timeout time.Duration, opts ...expect.Option) (*Context, error) {
	sshArgs := getSSHString(user, host)
	return (*spawner).Spawn(sshCommand, []string{sshArgs}, timeout, opts...)
}

func getSSHString(user, host string) string {
	return fmt.Sprintf("%s%s%s", user, sshSeparator, host)
}
