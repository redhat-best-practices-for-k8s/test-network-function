// Copyright (C) 2020-2022 Red Hat, Inc.
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

package configsections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	cID := ContainerIdentifier{
		NodeName:         "node1",
		Namespace:        "namespace1",
		PodName:          "pod1",
		ContainerName:    "container1",
		ContainerUID:     "uid1",
		ContainerRuntime: "runtime1",
	}
	assert.Equal(t, "node:node1 ns:namespace1 podName:pod1 containerName:container1 containerUID:uid1 containerRuntime:runtime1", cID.String())
}
