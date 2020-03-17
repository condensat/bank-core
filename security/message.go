// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"
)

var (
	ErrEncrypt              = errors.New("Failed to Encrypt")
	ErrOperationNotPermited = errors.New("Operation Not Permited")
	ErrSignature            = errors.New("Failed to Sign message")
	ErrNotSigned            = errors.New("Message Not Signed")
	ErrVerifyFailed         = errors.New("Fail to verify signature")
)

func SignMessage(ctx context.Context, key *Key, message *bank.Message) error {
	if key == nil {
		return ErrInvalidKey
	}

	if message == nil {
		return bank.ErrInvalidMessage
	}

	if message.IsSigned() {
		// NOOP
		return nil
	}

	if message.IsCompressed() || message.IsEncrypted() {
		return ErrOperationNotPermited
	}

	dataSig, err := key.SignMessage(ctx, message.Data)
	if err != nil {
		return ErrSignature
	}
	message.Data = dataSig[:]

	message.SetSigned(true)

	return nil
}

func VerifyMessageSignature(key SignaturePublicKey, message *bank.Message) (bool, error) {
	if message == nil {
		return false, bank.ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return false, bank.ErrNoData
	}

	if message.IsCompressed() || message.IsEncrypted() {
		return false, ErrOperationNotPermited
	}

	if !message.IsSigned() {
		return false, ErrNotSigned
	}

	valid, err := VerifySignature(key, message.Data)
	if err != nil {
		return false, ErrVerifyFailed
	}

	return valid, nil
}

func EncryptMessageFor(ctx context.Context, from *Key, to EncryptionPublicKey, message *bank.Message) error {
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return bank.ErrNoData
	}
	if message.IsEncrypted() {
		// NOOP
		return nil
	}

	data, err := from.EncryptFor(ctx, to, message.Data[:])
	if err != nil {
		return err
	}
	message.Data = data
	message.SetEncrypted(true)

	return nil
}

func DecryptMessageFrom(ctx context.Context, to *Key, from EncryptionPublicKey, message *bank.Message) error {
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return bank.ErrNoData
	}
	if !message.IsEncrypted() {
		// NOOP
		return nil
	}

	data, err := to.DecryptFrom(ctx, from, message.Data[:])
	if err != nil {
		return err
	}
	message.Data = data
	message.SetEncrypted(false)

	return nil
}
