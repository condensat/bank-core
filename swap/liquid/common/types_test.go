// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestProposalInfo_Args(t *testing.T) {
	t.Parallel()

	ref1 := ProposalInfo{
		ProposerAsset:  "assetP",
		ProposerAmount: 0.1234567811111,
		ReceiverAsset:  "assetR",
		ReceiverAmount: 3.141592653589793,
	}
	ref2 := ProposalInfo{
		ProposerAsset:  "assetP",
		ProposerAmount: -0.123456789,
		ReceiverAsset:  "assetR",
		ReceiverAmount: -3.141592653589793,
	}

	type fields struct {
		ProposalInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{"default", fields{}, []string{"", "0.00000000", "", "0.00000000"}},

		{"ref1", fields{ref1}, []string{"assetP", "0.12345678", "assetR", "3.14159265"}},
		{"ref2", fields{ref2}, []string{"assetP", "-0.12345679", "assetR", "-3.14159265"}},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposalInfo{
				ProposerAsset:  tt.fields.ProposerAsset,
				ProposerAmount: tt.fields.ProposerAmount,
				ReceiverAsset:  tt.fields.ReceiverAsset,
				ReceiverAmount: tt.fields.ReceiverAmount,
			}
			got := p.Args()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProposalInfo.Args() = %v, want %v", got, tt.want)
			}

			t.Logf("Args(), %v", got)
		})
	}
}

func TestPayload_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload Payload
		want    bool
	}{
		{"default", "", false},
		{"invalidJson", "invalid json", false},
		{"invalidBase64", Payload(base64.StdEncoding.EncodeToString([]byte("valid base64"))), false},

		{"rawJson", `{"valid": "json"}`, true},
		{"base64", Payload(base64.StdEncoding.EncodeToString([]byte(`{"valid": "json"}`))), true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.payload.Valid(); got != tt.want {
				t.Errorf("Payload.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
