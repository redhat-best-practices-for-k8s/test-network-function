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
	"os"
	"time"

	expect "github.com/ryandgoulding/goexpect"
)

const (
	shellEnvironmentVariableKey = "SHELL"
)

var (
	errorUnsetShell = fmt.Errorf("Environment variable SHELL is not set")
)

// SpawnShell creates an interactive shell subprocess based on the value of $SHELL, spawning the appropriate underlying
// PTY.
func SpawnShell(spawner *Spawner, timeout time.Duration, opts ...expect.Option) (*Context, error) {
	shellEnv, isSet := os.LookupEnv(shellEnvironmentVariableKey)
	if isSet != true {
		return nil, errorUnsetShell
	}
	var args []string
	return (*spawner).Spawn(shellEnv, args, timeout, opts...)
}
