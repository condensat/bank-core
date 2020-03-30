// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package internal

import (
	"testing"
)

func Test_lockKeyString(t *testing.T) {
	type args struct {
		prefix string
		value  interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Default", args{"", ""}, "lock."},
		{"Prefix", args{"foo", ""}, "foo."},
		{"PrefixValue", args{"foo", "bar"}, "foo.bar"},

		{"Int", args{"lock", 1}, "lock.1"},
		{"String", args{"lock", "foo"}, "lock.foo"},
		{"Float", args{"lock", 1.0}, "lock.1"},
		{"FloatDecimal", args{"lock", 1.1}, "lock.1.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lockKeyString(tt.args.prefix, tt.args.value); got != tt.want {
				t.Errorf("lockKeyString() = %v, want %v", got, tt.want)
			}
		})
	}
}
