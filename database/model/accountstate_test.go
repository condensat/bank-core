// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
)

func TestAccountStatus_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    AccountStatus
		want bool
	}{
		{"Default", AccountStatus(""), false},
		{"Invalid", AccountStatus("invalid"), false},
		{"NotValid", AccountStatus("not-valid"), false},

		{"Normal", AccountStatus("normal"), true},
		{"Locked", AccountStatus("locked"), true},
		{"Disabled", AccountStatus("disabled"), true},

		{"Random", AccountStatus("R4nD0m"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Valid(); got != tt.want {
				t.Errorf("AccountStatus.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAccountStatus(t *testing.T) {
	t.Parallel()

	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want AccountStatus
	}{
		{"Default", args{""}, AccountStatusInvalid},
		{"Invalid", args{"invalid"}, AccountStatusInvalid},
		{"NotValid", args{"not-valid"}, AccountStatusInvalid},

		{"Normal", args{"normal"}, AccountStatusNormal},
		{"Locked", args{"locked"}, AccountStatusLocked},
		{"Disabled", args{"disabled"}, AccountStatusDisabled},

		{"Random", args{"R4nD0m"}, AccountStatusInvalid},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseAccountStatus(tt.args.str); got != tt.want {
				t.Errorf("ParseAccountStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    AccountStatus
		want string
	}{
		{"Default", AccountStatus(""), ""},
		{"Invalid", AccountStatus("invalid"), ""},
		{"NotValid", AccountStatus("not-valid"), ""},

		{"Normal", AccountStatusNormal, "normal"},
		{"Locked", AccountStatusLocked, "locked"},
		{"Disabled", AccountStatusDisabled, "disabled"},

		{"Random", AccountStatus("R4nD0m"), ""},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("AccountStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_knownAccountStatus(t *testing.T) {
	t.Parallel()

	// do not use const here
	// keep order
	knownEnums := []string{
		"", // AccountStatusInvalid

		"normal",   // AccountStatusNormal
		"locked",   // AccountStatusAsyncLocked
		"disabled", // AccountStatusDisabled
	}
	var want []AccountStatus
	for _, enum := range knownEnums {
		want = append(want, ParseAccountStatus(enum))
	}

	if got := knownAccountStatus(); !reflect.DeepEqual(got, want) {
		t.Errorf("knownAccountStatus() = %v, want %v", got, want)
	}
}
