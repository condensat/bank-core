// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"testing"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/compression"
)

func TestSignMessage(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var zero [0]byte
	var data [32]byte

	message := bank.Message{
		Data: data[:],
	}
	messageZero := bank.Message{
		Data: zero[:],
	}
	sign := bank.Message{
		Data: data[:],
	}
	_ = SignMessage(shared, &sign)

	compress := bank.Message{
		Data: data[:],
	}
	_ = compression.CompressMessage(&compress, 5)

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encrypted)

	type args struct {
		key     bank.SharedKey
		message *bank.Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantSig bool
	}{
		{"nil", args{nil, nil}, true, false},
		{"nilkey", args{nil, &message}, true, false},
		{"nilmessage", args{shared, nil}, true, false},

		{"zero", args{nil, new(bank.Message)}, true, false},
		{"keyzero", args{shared, new(bank.Message)}, true, false},
		{"messagezero", args{shared, &messageZero}, true, false},
		{"compressed", args{shared, &compress}, true, false},
		{"encrypted", args{shared, &encrypted}, true, false},

		{"sign", args{shared, &message}, false, true},
		{"already_sign", args{shared, &sign}, false, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := SignMessage(tt.args.key, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("SignMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.message == nil {
				return
			}

			if tt.args.message.IsSigned() != tt.wantSig {
				t.Errorf("SignMessage() = %v, wantSig %v", tt.args.message.IsSigned(), tt.wantSig)
			}
		})
	}
}

func TestVerifyMessage(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var data [32]byte
	message := bank.Message{
		Data: data[:],
	}
	sign := bank.Message{
		Data: data[:],
	}
	_ = SignMessage(shared, &sign)
	wrongSign := bank.Message{
		Data: data[:],
	}
	_ = SignMessage(shared, &wrongSign)
	wrongSign.Signature = "NotAnHexaString"

	compress := bank.Message{
		Data: data[:],
	}
	_ = compression.CompressMessage(&compress, 5)

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encrypted)

	type args struct {
		key     bank.SharedKey
		message *bank.Message
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"nil", args{nil, nil}, false, true},
		{"nilkey", args{nil, &sign}, false, true},
		{"nilmessage", args{shared, nil}, false, true},

		{"zero", args{shared, new(bank.Message)}, false, true},
		{"compressed", args{shared, &compress}, false, true},
		{"encrypted", args{shared, &encrypted}, false, true},
		{"notsigned", args{shared, &message}, false, true},
		{"wrnongsign", args{shared, &wrongSign}, false, true},

		{"signed", args{shared, &sign}, true, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyMessage(tt.args.key, tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VerifyMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncryptMessageFor(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()

	var data [32]byte

	message := bank.Message{
		Data: data[:],
	}

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encrypted)

	type args struct {
		from    bank.PrivateKey
		to      bank.PublicKey
		message *bank.Message
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantEncrypt bool
	}{
		{"nil", args{nil, nil, nil}, true, false},
		{"nilmessage", args{priv, pub, nil}, true, false},
		{"nilkeys", args{nil, nil, new(bank.Message)}, true, false},
		{"nilkeysmessage", args{nil, nil, &message}, true, false},
		{"nilkeysencrypted", args{nil, nil, &encrypted}, true, true},
		{"encryptnodata", args{priv, pub, new(bank.Message)}, true, false},

		{"encrypt", args{priv, pub, &message}, false, true},
		{"encrypted", args{priv, pub, &encrypted}, false, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := EncryptMessageFor(tt.args.from, tt.args.to, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("EncryptMessageFor() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.message == nil {
				return
			}

			if tt.args.message.IsEncrypted() != tt.wantEncrypt {
				t.Errorf("EncryptMessageFor() = %v, wantEncrypt %v", tt.args.message.IsEncrypted(), tt.wantEncrypt)
			}
		})
	}
}

func TestEncryptMessage(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var data [32]byte

	message := bank.Message{
		Data: data[:],
	}

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encrypted)

	type args struct {
		key     bank.SharedKey
		message *bank.Message
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantEncrypt bool
	}{
		{"nil", args{nil, nil}, true, false},
		{"nilmessage", args{shared, nil}, true, false},
		{"nilkeys", args{shared, new(bank.Message)}, true, false},
		{"nilkeysmessage", args{nil, &message}, true, false},
		{"nilkeysencrypted", args{nil, &encrypted}, true, true},
		{"encryptnodata", args{shared, new(bank.Message)}, true, false},

		{"encrypt", args{shared, &message}, false, true},
		{"encrypted", args{shared, &encrypted}, false, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := EncryptMessage(tt.args.key, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("EncryptMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.message == nil {
				return
			}

			if tt.args.message.IsEncrypted() != tt.wantEncrypt {
				t.Errorf("EncryptMessage() = %v, wantEncrypt %v", tt.args.message.IsEncrypted(), tt.wantEncrypt)
			}
		})
	}
}

func TestDecryptMessageFrom(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()

	var data [32]byte

	message := bank.Message{
		Data: data[:],
	}

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encrypted)
	encryptedNoData := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encryptedNoData)
	encryptedNoData.Data = nil

	type args struct {
		to      bank.PrivateKey
		from    bank.PublicKey
		message *bank.Message
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantEncrypt bool
	}{
		{"nil", args{nil, nil, nil}, true, false},
		{"nilkeys", args{nil, nil, &message}, true, false},

		{"nilmessage", args{priv, pub, nil}, true, false},
		{"messagenodata", args{priv, pub, new(bank.Message)}, false, false},
		{"encryptednodata", args{priv, pub, &encryptedNoData}, true, true},

		{"notencrypted", args{priv, pub, &message}, false, false},
		{"encrypted", args{priv, pub, &encrypted}, false, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := DecryptMessageFrom(tt.args.to, tt.args.from, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("DecryptMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.message == nil {
				return
			}

			if tt.args.message.IsEncrypted() != tt.wantEncrypt {
				t.Errorf("DecryptMessage() = %v, wantEncrypt %v", tt.args.message.IsEncrypted(), tt.wantEncrypt)
			}
		})
	}
}

func TestDecryptMessage(t *testing.T) {
	t.Parallel()

	pub, priv, _ := NewKeys()
	shared, _ := SharedSecret(priv, pub)

	var data [32]byte

	message := bank.Message{
		Data: data[:],
	}

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encrypted)
	encryptedNoData := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(priv, pub, &encryptedNoData)
	encryptedNoData.Data = nil

	type args struct {
		key     bank.SharedKey
		message *bank.Message
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantEncrypt bool
	}{
		{"nil", args{nil, nil}, true, false},
		{"nilkeys", args{nil, &message}, true, false},

		{"nilmessage", args{shared, nil}, true, false},
		{"messagenodata", args{shared, new(bank.Message)}, false, false},
		{"encryptednodata", args{shared, &encryptedNoData}, true, true},

		{"notencrypted", args{shared, &message}, false, false},
		{"encrypted", args{shared, &encrypted}, false, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := DecryptMessage(tt.args.key, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("DecryptMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.message == nil {
				return
			}

			if tt.args.message.IsEncrypted() != tt.wantEncrypt {
				t.Errorf("DecryptMessage() = %v, wantEncrypt %v", tt.args.message.IsEncrypted(), tt.wantEncrypt)
			}
		})
	}
}
