// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"bytes"
	"testing"

	"github.com/condensat/bank-core"
)

func TestEncryptFor(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	type args struct {
		from bank.PrivateKey
		to   bank.PublicKey
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"nilpriv", args{nil, pub, data[:]}, 0, true},
		{"nilpub", args{priv, nil, data[:]}, 0, true},
		{"nilprivpub", args{nil, nil, data[:]}, 0, true},

		{"zeropriv", args{zero[:], pub, data[:]}, 0, true},
		{"zeropub", args{priv, zero[:], data[:]}, 0, true},
		{"zeropubpub", args{zero[:], zero[:], data[:]}, 0, true},

		{"nildata", args{priv, pub, nil}, 0, true},
		{"zerodata", args{priv, pub, zero[:]}, 0, true},

		{"allnil", args{nil, nil, nil}, 0, true},
		{"allzero", args{zero[:], zero[:], zero[:]}, 0, true},

		{"ecryptfor", args{priv, pub, data[:]}, 72, false},
		{"ecryptfor1", args{priv, pub, data1[:]}, 104, false},
		{"ecryptfor2", args{priv, pub, data2[:]}, 168, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptFor(tt.args.from, tt.args.to, tt.args.data)
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

func TestEncrypt(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	type args struct {
		sharedKey bank.SharedKey
		data      []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"nilshared", args{nil, data[:]}, 0, true},
		{"zeroshared", args{zero[:], data[:]}, 0, true},

		{"nildata", args{shared, nil}, 0, true},
		{"zerodata", args{shared, zero[:]}, 0, true},

		{"allnil", args{nil, nil}, 0, true},
		{"allzero", args{zero[:], zero[:]}, 0, true},

		{"ecrypt", args{shared, data[:]}, 72, false},
		{"ecrypt1", args{shared, data1[:]}, 104, false},
		{"ecrypt2", args{shared, data2[:]}, 168, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.args.sharedKey, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Encrypt() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestDecryptFrom(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	encrypt, _ := EncryptFor(priv, pub, data[:])
	encrypt1, _ := EncryptFor(priv, pub, data1[:])
	encrypt2, _ := EncryptFor(priv, pub, data2[:])

	type args struct {
		to   bank.PrivateKey
		from bank.PublicKey
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"nilpriv", args{nil, pub, data[:]}, nil, true},
		{"nilpub", args{priv, nil, data[:]}, nil, true},
		{"nilprivpub", args{nil, nil, data[:]}, nil, true},
		{"zeropriv", args{zero[:], pub, data[:]}, nil, true},
		{"zeropub", args{priv, zero[:], data[:]}, nil, true},
		{"zeropubpub", args{zero[:], zero[:], data[:]}, nil, true},

		{"nildata", args{priv, pub, nil}, nil, true},
		{"zerodata", args{priv, pub, zero[:]}, nil, true},

		{"allnil", args{nil, nil, nil}, nil, true},
		{"allzero", args{zero[:], zero[:], zero[:]}, nil, true},

		{"decryptfrom", args{priv, pub, encrypt[:]}, data[:], false},
		{"decryptfrom1", args{priv, pub, encrypt1[:]}, data1[:], false},
		{"decryptfrom2", args{priv, pub, encrypt2[:]}, data2[:], false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptFrom(tt.args.to, tt.args.from, tt.args.data)
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

func TestDecrypt(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	encrypt, _ := Encrypt(shared, data[:])
	encrypt1, _ := Encrypt(shared, data1[:])
	encrypt2, _ := Encrypt(shared, data2[:])

	type args struct {
		sharedKey bank.SharedKey
		data      []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"nilshared", args{nil, data[:]}, nil, true},
		{"zeroshared", args{zero[:], data[:]}, nil, true},

		{"nildata", args{shared, nil}, nil, true},
		{"zerodata", args{shared, zero[:]}, nil, true},

		{"allnil", args{nil, nil}, nil, true},
		{"allzero", args{zero[:], zero[:]}, nil, true},

		{"decrypt", args{shared, encrypt[:]}, data[:], false},
		{"decrypt1", args{shared, encrypt1[:]}, data1[:], false},
		{"decrypt2", args{shared, encrypt2[:]}, data2[:], false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.args.sharedKey, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
