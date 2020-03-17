// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"context"
	"testing"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/compression"
	"github.com/condensat/bank-core/security/utils"
)

func TestSignMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	key := NewKey(ctx)

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
	_ = SignMessage(ctx, key, &sign)

	compress := bank.Message{
		Data: data[:],
	}
	_ = compression.CompressMessage(&compress, 5)

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(ctx, key, key.Public(ctx), &encrypted)

	type args struct {
		key     *Key
		message *bank.Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantSig bool
	}{
		{"nilmessage", args{key, nil}, true, false},

		{"keyzero", args{key, new(bank.Message)}, true, false},
		{"messagezero", args{key, &messageZero}, true, false},
		{"compressed", args{key, &compress}, true, false},
		{"encrypted", args{key, &encrypted}, true, false},

		{"sign", args{key, &message}, false, true},
		{"already_sign", args{key, &sign}, false, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := SignMessage(ctx, tt.args.key, tt.args.message); (err != nil) != tt.wantErr {
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

func TestVerifyMessageSignature(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	key := NewKey(ctx)

	var data [32]byte
	message := bank.Message{
		Data: data[:],
	}
	sign := bank.Message{
		Data: data[:],
	}
	_ = SignMessage(ctx, key, &sign)

	compress := bank.Message{
		Data: data[:],
	}
	_ = compression.CompressMessage(&compress, 5)

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(ctx, key, key.Public(ctx), &encrypted)

	type args struct {
		key     *Key
		message *bank.Message
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"nilmessage", args{key, nil}, false, true},

		{"zero", args{key, new(bank.Message)}, false, true},
		{"compressed", args{key, &compress}, false, true},
		{"encrypted", args{key, &encrypted}, false, true},
		{"notsigned", args{key, &message}, false, true},

		{"signed", args{key, &sign}, true, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyMessageSignature(tt.args.key.SignPublicKey(ctx), tt.args.message)
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

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	key := NewKey(ctx)

	var data [32]byte

	message := bank.Message{
		Data: data[:],
	}

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(ctx, key, key.Public(ctx), &encrypted)

	type args struct {
		from    *Key
		message *bank.Message
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantEncrypt bool
	}{
		{"nilmessage", args{key, nil}, true, false},
		{"encryptnodata", args{key, new(bank.Message)}, true, false},

		{"encrypt", args{key, &message}, false, true},
		{"encrypted", args{key, &encrypted}, false, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := EncryptMessageFor(ctx, tt.args.from, tt.args.from.Public(ctx), tt.args.message); (err != nil) != tt.wantErr {
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

func TestDecryptMessageFrom(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyPrivateKeySalt, utils.GenerateRandN(32))
	key := NewKey(ctx)

	var data [32]byte

	message := bank.Message{
		Data: data[:],
	}

	encrypted := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(ctx, key, key.Public(ctx), &encrypted)
	encryptedNoData := bank.Message{
		Data: data[:],
	}
	_ = EncryptMessageFor(ctx, key, key.Public(ctx), &encryptedNoData)
	encryptedNoData.Data = nil

	type args struct {
		to      *Key
		message *bank.Message
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantEncrypt bool
	}{
		{"nilmessage", args{key, nil}, true, false},
		{"messagenodata", args{key, new(bank.Message)}, true, false},
		{"encryptednodata", args{key, &encryptedNoData}, true, true},

		{"notencrypted", args{key, &message}, false, false},
		{"encrypted", args{key, &encrypted}, false, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := DecryptMessageFrom(ctx, tt.args.to, tt.args.to.Public(ctx), tt.args.message); (err != nil) != tt.wantErr {
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
