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

package generic

// Match follows the Container design pattern, and is used to store the arguments to a reel.Handler's ReelMatch
// function in a single data transfer object.
type Match struct {

	// Pattern is the pattern causing a match in reel.Handler ReelMatch.
	Pattern string `json:"pattern,omitempty" yaml:"pattern,omitempty"`

	// Before contains all output preceding Match.
	Before string `json:"before,omitempty" yaml:"before,omitempty"`

	// Match is the matched string.
	Match string `json:"match,omitempty" yaml:"match,omitempty"`
}
