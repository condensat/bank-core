// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"
)

func TestCryptoAddress_IsUsed(t *testing.T) {
	t.Parallel()

	type fields struct {
		FirstBlockId BlockID
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, false},
		{"mempool", fields{BlockID(1)}, true},
		{"mined", fields{BlockID(424242)}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &CryptoAddress{
				FirstBlockId: tt.fields.FirstBlockId,
			}
			if got := p.IsUsed(); got != tt.want {
				t.Errorf("CryptoAddress.IsUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCryptoAddress_Confirmations(t *testing.T) {
	t.Parallel()

	type fields struct {
		FirstBlockId BlockID
	}
	type args struct {
		height BlockID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{"default", fields{}, args{}, 0},
		{"mempool", fields{BlockID(1)}, args{424242}, 0},
		{"mined", fields{BlockID(424242)}, args{424242}, 1},
		{"confirmed", fields{BlockID(424242)}, args{424247}, 6},
		{"future", fields{BlockID(424243)}, args{424242}, 0},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &CryptoAddress{
				FirstBlockId: tt.fields.FirstBlockId,
			}
			if got := p.Confirmations(tt.args.height); got != tt.want {
				t.Errorf("CryptoAddress.Confirmations() = %v, want %v", got, tt.want)
			}
		})
	}
}
