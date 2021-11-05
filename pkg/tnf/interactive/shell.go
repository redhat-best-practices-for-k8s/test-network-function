// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"log"
	"os"
	"time"
)

const (
	shellEnvironmentVariableKey = "SHELL"
	defaultTimeoutSeconds       = 10
	defaultTimeout              = defaultTimeoutSeconds * time.Second
)

// SpawnShell creates an interactive shell subprocess based on the value of $SHELL, spawning the appropriate underlying
// PTY.
func SpawnShell(spawner *Spawner, timeout time.Duration, opts ...Option) (*Context, error) {
	shellEnv := os.Getenv(shellEnvironmentVariableKey)
	var args []string
	return (*spawner).Spawn(shellEnv, args, timeout, opts...)
}

//
//
// GetContext spawns a new shell session and returns its context
func GetContext(verbose bool) *Context {
	context, err := SpawnShell(CreateGoExpectSpawner(), defaultTimeout, Verbose(verbose), SendTimeout(defaultTimeout))
	if err != nil || context == nil || context.GetExpecter() == nil {
		log.Panicf("can't get a proper context for test execution")
	}
	return context
}
