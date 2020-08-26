// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "testing"

func TestSsmAddress_Valid(t *testing.T) {
	type fields struct {
		ID            SsmAddressID
		PublicAddress SsmPublicAddress
		ScriptPubkey  SsmPubkey
		BlindingKey   SsmBlindingKey
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default", fields{}, false},
		{"InvalidPublicAddress", fields{0, "", "bar", ""}, false},
		{"InvalidScriptPubkey", fields{0, "foo", "", ""}, false},
		{"InvalidWithOptional", fields{0, "", "", "foobar"}, false},
		{"InvalidPublicAddressOptional", fields{0, "", "bar", "foobar"}, false},
		{"InvalidScriptPubkeyOptional", fields{0, "foo", "", "foobar"}, false},

		{"Valid", fields{0, "foo", "bar", ""}, true},
		{"Optional", fields{0, "foo", "bar", "foobar"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SsmAddress{
				ID:            tt.fields.ID,
				PublicAddress: tt.fields.PublicAddress,
				ScriptPubkey:  tt.fields.ScriptPubkey,
				BlindingKey:   tt.fields.BlindingKey,
			}
			if got := p.IsValid(); got != tt.want {
				t.Errorf("SsmAddress.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
