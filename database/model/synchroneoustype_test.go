// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
)

func TestSynchroneousType_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    SynchroneousType
		want bool
	}{
		{"Default", SynchroneousType(""), false},
		{"Invalid", SynchroneousType("invalid"), false},
		{"NotValid", SynchroneousType("not-valid"), false},

		{"Sync", SynchroneousType("sync"), true},
		{"AsyncStart", SynchroneousType("async-start"), true},
		{"AsyncEnd", SynchroneousType("async-end"), true},

		{"Random", SynchroneousType("R4nD0m"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Valid(); got != tt.want {
				t.Errorf("SynchroneousType.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSynchroneousType(t *testing.T) {
	t.Parallel()

	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want SynchroneousType
	}{
		{"Default", args{""}, SynchroneousTypeInvalid},
		{"Invalid", args{"invalid"}, SynchroneousTypeInvalid},
		{"NotValid", args{"not-valid"}, SynchroneousTypeInvalid},

		{"Sync", args{"sync"}, SynchroneousTypeSync},
		{"AsyncStart", args{"async-start"}, SynchroneousTypeAsyncStart},
		{"AsyncEnd", args{"async-end"}, SynchroneousTypeAsyncEnd},

		{"Random", args{"R4nD0m"}, SynchroneousTypeInvalid},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSynchroneousType(tt.args.str); got != tt.want {
				t.Errorf("ParseSynchroneousType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSynchroneousType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    SynchroneousType
		want string
	}{
		{"Default", SynchroneousType(""), ""},
		{"Invalid", SynchroneousType("invalid"), ""},
		{"NotValid", SynchroneousType("not-valid"), ""},

		{"Sync", SynchroneousTypeSync, "sync"},
		{"AsyncStart", SynchroneousTypeAsyncStart, "async-start"},
		{"AsyncEnd", SynchroneousTypeAsyncEnd, "async-end"},

		{"Random", SynchroneousType("R4nD0m"), ""},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("SynchroneousType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_knownSynchroneousType(t *testing.T) {
	t.Parallel()

	// do not use const here
	// keep order
	knownEnums := []string{
		"", // SynchroneousTypeInvalid

		"sync",        // SynchroneousTypeSync
		"async-start", // SynchroneousTypeAsyncStart
		"async-end",   // SynchroneousTypeAsyncEnd
	}
	var want []SynchroneousType
	for _, enum := range knownEnums {
		want = append(want, ParseSynchroneousType(enum))
	}

	if got := knownSynchroneousType(); !reflect.DeepEqual(got, want) {
		t.Errorf("knownSynchroneousType() = %v, want %v", got, want)
	}
}
