// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"bytes"
	"testing"

	"github.com/condensat/bank-core"
)

func TestIsKeyValid(t *testing.T) {
	t.Parallel()

	type args struct {
		key []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil", args{nil}, false},
		{"zero", args{new([0]byte)[:]}, false},
		{"one", args{new([1]byte)[:]}, false},
		{"ok", args{new([32]byte)[:]}, true},
		{"ko", args{new([64]byte)[:]}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKeyValid(tt.args.key); got != tt.want {
				t.Errorf("IsKeyValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKeys(t *testing.T) {
	t.Parallel()

	var zero [32]byte
	if !IsKeyValid(zero[:]) {
		t.Errorf("Invalid zero key")
	}
	tests := []struct {
		name    string
		want    bool
		want1   bool
		wantErr bool
	}{
		{"new", true, true, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := NewKeys()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if IsKeyValid(got[:]) != tt.want {
				t.Errorf("NewKeys() got = %v, want %v", got, tt.want)
			}
			if bytes.Equal(got[:], zero[:]) {
				t.Errorf("NewKeys() got is a zero key")
			}

			if IsKeyValid(got1[:]) != tt.want1 {
				t.Errorf("NewKeys() got1 = %v, want %v", got1, tt.want1)
			}
			if bytes.Equal(got1[:], zero[:]) {
				t.Errorf("NewKeys() got1 is a zero key")
			}

			if bytes.Equal(got[:], got1[:]) {
				t.Errorf("NewKeys() got same public and private key")
			}
		})
	}
}

func TestSharedSecret(t *testing.T) {
	t.Parallel()

	var zero []byte = nil
	pub, priv, _ := NewKeys()
	type args struct {
		privateKey bank.PrivateKey
		publicKey  bank.PublicKey
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"shared", args{priv, pub}, true, false},
		{"zero_priv", args{bank.PrivateKey(zero[:]), pub}, false, true},
		{"zero_pub", args{bank.PrivateKey(priv), zero[:]}, false, true},
		{"zero_priv_pub", args{bank.PrivateKey(zero[:]), zero[:]}, false, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := SharedSecret(tt.args.privateKey, tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("SharedSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if IsKeyValid(got[:]) != tt.want {
				t.Errorf("SharedSecret() = %v, want %v", IsKeyValid(got), tt.want)
			}
		})
	}
}
