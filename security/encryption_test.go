// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"bytes"
	"context"
	"testing"

	"github.com/condensat/bank-core/security/utils"
)

func TestEncryptFor(t *testing.T) {
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
		{"zerodata", args{key, zero[:]}, 0, true},

		{"ecryptfor", args{key, data[:]}, 72, false},
		{"ecryptfor1", args{key, data1[:]}, 104, false},
		{"ecryptfor2", args{key, data2[:]}, 168, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.key.EncryptFor(ctx, tt.args.key.Public(ctx), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("EncryptFor() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestDecryptFrom(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	key := NewKey(ctx)

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	encrypt, _ := key.EncryptFor(ctx, key.Public(ctx), data[:])
	encrypt1, _ := key.EncryptFor(ctx, key.Public(ctx), data1[:])
	encrypt2, _ := key.EncryptFor(ctx, key.Public(ctx), data2[:])

	type args struct {
		key  *Key
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"nilpub", args{key, data[:]}, nil, true},

		{"nildata", args{key, nil}, nil, true},
		{"zerodata", args{key, zero[:]}, nil, true},

		{"decryptfrom", args{key, encrypt[:]}, data[:], false},
		{"decryptfrom1", args{key, encrypt1[:]}, data1[:], false},
		{"decryptfrom2", args{key, encrypt2[:]}, data2[:], false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.key.DecryptFrom(ctx, tt.args.key.Public(ctx), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("DecryptFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
