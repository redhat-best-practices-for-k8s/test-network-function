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

package common

import (
	"github.com/onsi/ginkgo"
	log "github.com/sirupsen/logrus"
	configpkg "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/autodiscover"
)

var env *configpkg.TestEnvironment

var _ = ginkgo.BeforeSuite(func() {
	for name := range autodiscover.GetNodesList() {
		autodiscover.DeleteDebugLabel(name)
	}
})

var _ = ginkgo.AfterSuite(func() {
	// clean up added label to nodes
	log.Info("clean up added labels to nodes")
	env = configpkg.GetTestEnvironment()
	env.LoadAndRefresh()
	for name, node := range env.NodesUnderTest {
		if !(node.HasDebugPod()) {
			continue
		}
		node.Oc.Close()
		autodiscover.DeleteDebugLabel(name)
	}
})
