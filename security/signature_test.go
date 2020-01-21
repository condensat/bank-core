// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"testing"

	"github.com/condensat/bank-core"
)

func TestSign(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	type args struct {
		key  bank.SharedKey
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"allnil", args{nil, nil}, 0, true},
		{"nilkey", args{nil, data[:]}, 0, true},
		{"nildata", args{shared, nil}, 0, true},
		{"zero", args{shared, zero[:]}, 0, true},

		{"data", args{shared, data[:]}, 64, false},
		{"data1", args{shared, data1[:]}, 64, false},
		{"data2", args{shared, data2[:]}, 64, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sign(tt.args.key, tt.args.data)
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

func TestVerify(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	sig, _ := Sign(shared, data[:])
	sig1, _ := Sign(shared, data1[:])
	sig2, _ := Sign(shared, data2[:])

	type args struct {
		key       bank.SharedKey
		data      []byte
		signature []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"allnil", args{nil, nil, nil}, false},
		{"nilkey", args{nil, data[:], sig[:]}, false},
		{"nildata", args{shared[:], nil, sig[:]}, false},
		{"nilsig", args{shared[:], data[:], nil}, false},
		{"zero", args{shared[:], zero[:], sig[:]}, false},

		{"sig", args{shared[:], data[:], sig[:]}, true},
		{"sig1", args{shared[:], data1[:], sig1[:]}, true},
		{"sig2", args{shared[:], data2[:], sig2[:]}, true},

		{"datasig1", args{shared[:], data[:], sig1[:]}, false},
		{"datasig2", args{shared[:], data[:], sig2[:]}, false},
		{"data1sig", args{shared[:], data1[:], sig[:]}, false},
		{"data1sig2", args{shared[:], data1[:], sig2[:]}, false},
		{"data2sig", args{shared[:], data2[:], sig[:]}, false},
		{"data2sig1", args{shared[:], data2[:], sig1[:]}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := Verify(tt.args.key, tt.args.data, tt.args.signature); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
