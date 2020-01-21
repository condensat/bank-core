// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"crypto/rand"
	"errors"

	"github.com/condensat/bank-core"

	"golang.org/x/crypto/nacl/box"
)

var (
	ErrGenerateKey = errors.New("Failed to generate keys")
	ErrInvalidKey  = errors.New("Invalid key")
)

func NewKeys() (bank.PublicKey, bank.PrivateKey, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return bank.PublicKey(nil), bank.PrivateKey(nil), ErrGenerateKey
	}

	if len(publicKey) == 0 || len(privateKey) == 0 {
		return bank.PublicKey(nil), bank.PrivateKey(nil), ErrGenerateKey
	}

	return bank.PublicKey(publicKey[:]),
		bank.PrivateKey(privateKey[:]),
		nil
}

func IsKeyValid(key []byte) bool {
	return len(key) == 32
}

func SharedSecret(privateKey bank.PrivateKey, publicKey bank.PublicKey) (bank.SharedKey, error) {
	if !IsKeyValid(privateKey[:]) || !IsKeyValid(publicKey[:]) {
		return bank.SharedKey(nil), ErrInvalidKey
	}

	var private [32]byte
	var public [32]byte
	copy(private[:], privateKey[:])
	copy(public[:], publicKey[:])

	var shared [32]byte
	box.Precompute(&shared, &public, &private)

	return bank.SharedKey(shared[:]), nil
}
