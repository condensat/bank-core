// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"encoding/hex"
	"errors"

	"github.com/condensat/bank-core"
)

var (
	ErrEncrypt              = errors.New("Failed to Encrypted")
	ErrOperationNotPermited = errors.New("Operation Not Permited")
	ErrSignature            = errors.New("Failed to Sign message")
	ErrNotSigned            = errors.New("Message Not Signed")
)

func SignMessage(key bank.SharedKey, message *bank.Message) error {
	if !IsKeyValid(key) {
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

	sig, err := Sign(key, message.Data)
	if err != nil {
		return ErrSignature
	}

	message.Signature = hex.EncodeToString(sig)
	message.SetSigned(true)

	return nil

}

func VerifyMessage(key bank.SharedKey, message *bank.Message) (bool, error) {
	if !IsKeyValid(key) {
		return false, ErrInvalidKey
	}
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

	sig, err := hex.DecodeString(message.Signature)
	if err != nil {
		return false, ErrNotSigned
	}

	return Verify(key, message.Data, sig), nil
}

func EncryptMessageFor(from bank.PrivateKey, to bank.PublicKey, message *bank.Message) error {
	if !IsKeyValid(from) || !IsKeyValid(to) {
		return ErrInvalidKey
	}
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if message.IsEncrypted() {
		// NOOP
		return nil
	}

	data, err := EncryptFor(from, to, message.Data[:])
	if err != nil {
		return err
	}
	message.Data = data
	message.SetEncrypted(true)

	return nil
}

func EncryptMessage(key bank.SharedKey, message *bank.Message) error {
	if !IsKeyValid(key) {
		return ErrInvalidKey
	}
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if message.IsEncrypted() {
		// NOOP
		return nil
	}

	data, err := Encrypt(key, message.Data[:])
	if err != nil {
		return err
	}
	message.Data = data
	message.SetEncrypted(true)

	return nil
}

func DecryptMessageFrom(to bank.PrivateKey, from bank.PublicKey, message *bank.Message) error {
	if !IsKeyValid(to) || !IsKeyValid(from) {
		return ErrInvalidKey
	}
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if !message.IsEncrypted() {
		// NOOP
		return nil
	}

	data, err := DecryptFrom(to, from, message.Data[:])
	if err != nil {
		return err
	}
	message.Data = data
	message.SetEncrypted(false)

	return nil
}

func DecryptMessage(key bank.SharedKey, message *bank.Message) error {
	if !IsKeyValid(key) {
		return ErrInvalidKey
	}
	if message == nil {
		return bank.ErrInvalidMessage
	}
	if !message.IsEncrypted() {
		// NOOP
		return nil
	}

	data, err := Decrypt(key, message.Data[:])
	if err != nil {
		return err
	}
	message.Data = data
	message.SetEncrypted(false)

	return nil
}
