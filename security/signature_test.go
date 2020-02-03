// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"context"
	"testing"

	"github.com/condensat/bank-core/security/utils"
)

func TestSign(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	key := NewKey(ctx)

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	type args struct {
		key  *Key
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"nildata", args{key, nil}, 0, true},
		{"zero", args{key, zero[:]}, 0, true},

		{"data", args{key, data[:]}, 96, false},
		{"data1", args{key, data1[:]}, 128, false},
		{"data2", args{key, data2[:]}, 192, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {

			privateKey := tt.args.key.privateKey(ctx)
			signatureKey := SignatureSecretKey(privateKey)

			got, err := Sign(signatureKey, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Sign() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	key := NewKey(ctx)

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	dataSig, _ := key.SignMessage(ctx, data[:])
	dataSig1, _ := key.SignMessage(ctx, data1[:])
	dataSig2, _ := key.SignMessage(ctx, data2[:])

	type args struct {
		key  *Key
		data []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nildata", args{key, nil}, false},
		{"nilsig", args{key, data[:]}, false},
		{"zero", args{key, zero[:]}, false},

		{"sig", args{key, dataSig[:]}, true},
		{"sig1", args{key, dataSig1[:]}, true},
		{"sig2", args{key, dataSig2[:]}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := VerifySignature(tt.args.key.SignPublicKey(ctx), tt.args.data); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
