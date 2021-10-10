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

package configsections

// WorkerLabel const for k8s worker
const WorkerLabel = "node-role.kubernetes.io/worker"

// MasterLabel const for k8s for master
const MasterLabel = "node-role.kubernetes.io/master"

// Node defines in the cluster. with name of the node and the type of this node master/worker,,,,.
type Node struct {
	Name   string
	Labels []string
}

// IsMaster Function that return if the node is master
func (node Node) IsMaster() bool {
	for _, t := range node.Labels {
		if t == MasterLabel {
			return true
		}
	}
	return false
}

// IsWorker Function that return if the node is worker
func (node Node) IsWorker() bool {
	for _, t := range node.Labels {
		if t == WorkerLabel {
			return true
		}
	}
	return false
}
