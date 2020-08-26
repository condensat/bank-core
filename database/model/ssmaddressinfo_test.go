// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "testing"

func TestSsmAddressInfo_IsValid(t *testing.T) {
	type fields struct {
		SsmAddressID SsmAddressID
		Chain        SsmChain
		Fingerprint  SsmFingerprint
		HDPath       SsmHDPath
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default", fields{}, false},
		{"InvalidID", fields{0, "chain", "fingerprint", "path"}, false},
		{"InvalidChain", fields{42, "", "fingerprint", "path"}, false},
		{"InvalidFingerprint", fields{42, "chain", "", "path"}, false},
		{"InvalidPath", fields{42, "chain", "fingerprint", ""}, false},

		{"Valid", fields{42, "chain", "fingerprint", "path"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SsmAddressInfo{
				SsmAddressID: tt.fields.SsmAddressID,
				Chain:        tt.fields.Chain,
				Fingerprint:  tt.fields.Fingerprint,
				HDPath:       tt.fields.HDPath,
			}
			if got := p.IsValid(); got != tt.want {
				t.Errorf("SsmAddressInfo.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
