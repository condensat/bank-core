// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"context"
	"errors"

	"github.com/condensat/bank-core/security"
)

var (
	ErrEncrypt      = errors.New("Failed to Encrypt")
	ErrSignature    = errors.New("Failed to Sign message")
	ErrNotSigned    = errors.New("Message Not Signed")
	ErrVerifyFailed = errors.New("Fail to verify signature")
)

func SignMessage(ctx context.Context, key *security.Key, message *Message) error {
	if key == nil {
		return security.ErrInvalidKey
	}

	if message == nil {
		return ErrInvalidMessage
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

func VerifyMessageSignature(key security.SignaturePublicKey, message *Message) (bool, error) {
	if message == nil {
		return false, ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return false, ErrNoData
	}

	if message.IsCompressed() || message.IsEncrypted() {
		return false, ErrOperationNotPermited
	}

	if !message.IsSigned() {
		return false, ErrNotSigned
	}

	valid, err := security.VerifySignature(key, message.Data)
	if err != nil {
		return false, ErrVerifyFailed
	}

	return valid, nil
}

func EncryptMessageFor(ctx context.Context, from *security.Key, to security.EncryptionPublicKey, message *Message) error {
	if message == nil {
		return ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return ErrNoData
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

func DecryptMessageFrom(ctx context.Context, to *security.Key, from security.EncryptionPublicKey, message *Message) error {
	if message == nil {
		return ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return ErrNoData
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
