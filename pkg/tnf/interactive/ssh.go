// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package interactive

import (
	"fmt"
	"time"

	expect "github.com/google/goexpect"
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
