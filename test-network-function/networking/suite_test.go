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

package networking

import (
	"reflect"
	"testing"
)

func TestIsIPv4AndIsIPv6(t *testing.T) {
	type args struct {
		address string
	}
	type test struct {
		name string
		args args
		want bool
	}
	// Passing tests
	var tests = []test{
		{
			name: "ipv4",
			args: args{address: "2.2.2.2"},
			want: true,
		}, {
			name: "ipv6",
			args: args{address: "fd00:10:244:1::3"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// testing IsIPv4
			if got := IsIPv4(tt.args.address); got != tt.want {
				t.Errorf("IsIPv4() = %v, want %v", got, tt.want)
			}
			// testing IsIPv6
			if got := IsIPv6(tt.args.address); got == tt.want {
				t.Errorf("IsIPv6() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterIPListPerVersion(t *testing.T) {
	type args struct {
		ipList    []string
		ipVersion string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "passingIPv4",
			args: args{ipList: []string{"2.2.2.2", "3.3.3.3", "fd00:10:244:1::3", "fd00:10:244:1::4"}, ipVersion: IPv4},
			want: []string{"2.2.2.2", "3.3.3.3"},
		},
		{name: "passingIPv6",
			args: args{ipList: []string{"2.2.2.2", "3.3.3.3", "fd00:10:244:1::3", "fd00:10:244:1::4"}, ipVersion: IPv6},
			want: []string{"fd00:10:244:1::3", "fd00:10:244:1::4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterIPListPerVersion(tt.args.ipList, tt.args.ipVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterIPListPerVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
