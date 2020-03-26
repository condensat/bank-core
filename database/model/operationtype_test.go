// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
)

func TestOperationType_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    OperationType
		want bool
	}{
		{"Default", OperationType(""), false},
		{"Invalid", OperationType("invalid"), false},
		{"NotValid", OperationType("not-valid"), false},

		{"Deposit", OperationType("deposit"), true},
		{"Withdraw", OperationType("withdraw"), true},
		{"Transfert", OperationType("transfert"), true},

		{"None", OperationType("none"), true},
		{"Other", OperationType("other"), true},

		{"Random", OperationType("R4nD0m"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Valid(); got != tt.want {
				t.Errorf("OperationType.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOperationType(t *testing.T) {
	t.Parallel()

	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want OperationType
	}{
		{"Default", args{""}, OperationTypeInvalid},
		{"Invalid", args{"invalid"}, OperationTypeInvalid},
		{"NotValid", args{"not-valid"}, OperationTypeInvalid},

		{"Deposit", args{"deposit"}, OperationTypeDeposit},
		{"Withdraw", args{"withdraw"}, OperationTypeWithdraw},
		{"Transfert", args{"transfert"}, OperationTypeTransfert},
		{"Adjustment", args{"adjustment"}, OperationTypeAdjustment},

		{"None", args{"none"}, OperationTypeNone},
		{"Other", args{"other"}, OperationTypeOther},

		{"Random", args{"R4nD0m"}, OperationTypeInvalid},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseOperationType(tt.args.str); got != tt.want {
				t.Errorf("ParseOperationType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperationType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    OperationType
		want string
	}{
		{"Default", OperationType(""), ""},
		{"Invalid", OperationType("invalid"), ""},
		{"NotValid", OperationType("not-valid"), ""},

		{"Deposit", OperationType("deposit"), "deposit"},
		{"Withdraw", OperationType("withdraw"), "withdraw"},
		{"Transfert", OperationType("transfert"), "transfert"},
		{"Adjustment", OperationType("adjustment"), "adjustment"},

		{"None", OperationType("none"), "none"},
		{"Other", OperationType("other"), "other"},

		{"Random", OperationType("R4nD0m"), ""},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("OperationType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_knownOperationType(t *testing.T) {
	t.Parallel()

	// do not use const here
	// keep order
	knownEnums := []string{
		"", //OperationTypeInvalid

		"deposit",    // OperationTypeDeposit
		"withdraw",   //OperationTypeWithdraw
		"transfert",  // OperationTypeTransfert
		"adjustment", // OperationTypeAdjustment

		"none",  // OperationTypeNone
		"other", // OperationTypeOther
	}
	var want []OperationType
	for _, enum := range knownEnums {
		want = append(want, ParseOperationType(enum))
	}

	if got := knownOperationType(); !reflect.DeepEqual(got, want) {
		t.Errorf("knownOperationType() = %v, want %v", got, want)
	}
}
