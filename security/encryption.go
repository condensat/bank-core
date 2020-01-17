// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"crypto/rand"
	"errors"
	"io"

	"github.com/condensat/bank-core"

	"golang.org/x/crypto/nacl/box"
)

var (
	ErrNonce        = errors.New("Nonce Error")
	ErrInvalidData  = errors.New("InvalidData Error")
	ErrSharedSecret = errors.New("SharedSecret Error")
	ErrDecrypt      = errors.New("Decrypt Error")
)

func EncryptFor(from bank.PrivateKey, to bank.PublicKey, data []byte) ([]byte, error) {
	if !IsKeyValid(from[:]) || !IsKeyValid(to[:]) {
		return nil, ErrSharedSecret
	}
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	sharedKey, err := SharedSecret(from, to)
	if err != nil {
		return nil, ErrSharedSecret

	}

	return Encrypt(sharedKey, data)
}

func Encrypt(sharedKey bank.SharedKey, data []byte) ([]byte, error) {
	if !IsKeyValid(sharedKey[:]) {
		return nil, ErrSharedSecret
	}
	if len(data) == 0 {
		return nil, ErrInvalidData
	}
	var shared [32]byte
	copy(shared[:], sharedKey[:])

	var nonce [24]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, ErrNonce
	}

	return box.SealAfterPrecomputation(nonce[:], []byte(data), &nonce, &shared), nil
}

func DecryptFrom(to bank.PrivateKey, from bank.PublicKey, data []byte) ([]byte, error) {
	if !IsKeyValid(from[:]) || !IsKeyValid(to[:]) {
		return nil, ErrSharedSecret
	}
	if len(data) == 0 {
		return nil, ErrInvalidData
	}

	sharedKey, err := SharedSecret(to, from)
	if err != nil {
		return nil, ErrSharedSecret

	}

	return Decrypt(sharedKey, data)
}

func Decrypt(sharedKey bank.SharedKey, data []byte) ([]byte, error) {
	if !IsKeyValid(sharedKey[:]) {
		return nil, ErrSharedSecret
	}
	if len(data) == 0 {
		return nil, ErrInvalidData
	}
	var shared [32]byte
	copy(shared[:], sharedKey[:])

	var nonce [24]byte
	copy(nonce[:], data[:24])
	data, ok := box.OpenAfterPrecomputation(nil, data[24:], &nonce, &shared)
	if !ok {
		return nil, ErrDecrypt
	}

	return data, nil
}
