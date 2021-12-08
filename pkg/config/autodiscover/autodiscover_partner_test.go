// Copyright (C) 2021 Red Hat, Inc.
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

package autodiscover

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNonOcpCluster(t *testing.T) {
	defer os.Unsetenv("TNF_NON_OCP_CLUSTER")
	testCases := []struct {
		IsNonOcpCluster bool
	}{
		{
			IsNonOcpCluster: true,
		},
		{
			IsNonOcpCluster: false,
		},
	}

	for _, tc := range testCases {
		if tc.IsNonOcpCluster {
			os.Setenv("TNF_NON_OCP_CLUSTER", "true")
			assert.True(t, IsNonOcpCluster())
		} else {
			os.Setenv("TNF_NON_OCP_CLUSTER", "false")
			assert.False(t, IsNonOcpCluster())
		}
	}
}
