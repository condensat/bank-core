// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "testing"

func TestAssetInfo_Valid(t *testing.T) {
	type fields struct {
		AssetID   AssetID
		Domain    string
		Name      string
		Ticker    string
		Precision uint8
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, false},
		{"invalidDomain", fields{42, "", "FooBar", "FBAR", 0}, false},
		{"invalidName", fields{42, "foo.bar", "", "FBAR", 0}, false},
		{"invalidTicker", fields{42, "foo.bar", "FooBar", "", 0}, false},
		{"invalidTicker", fields{42, "foo.bar", "FooBar", "F00BAR", 0}, false},

		{"valid", fields{42, "foo.bar", "FooBar", "FBAR", 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &AssetInfo{
				AssetID:   tt.fields.AssetID,
				Domain:    tt.fields.Domain,
				Name:      tt.fields.Name,
				Ticker:    tt.fields.Ticker,
				Precision: tt.fields.Precision,
			}
			if got := p.Valid(); got != tt.want {
				t.Errorf("AssetInfo.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
