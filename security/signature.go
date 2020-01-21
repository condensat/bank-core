// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"crypto/ed25519"
	"crypto/sha256"

	"github.com/condensat/bank-core"
)

func Sign(key bank.SharedKey, data []byte) ([]byte, error) {
	if !IsKeyValid(key) {
		return nil, ErrInvalidKey
	}
	if len(data) == 0 {
		return nil, bank.ErrNoData
	}

	hash := sha256.Sum256(key[:])
	priv := ed25519.NewKeyFromSeed(hash[:])

	return ed25519.Sign(priv, data), nil
}

func Verify(key bank.SharedKey, data, signature []byte) bool {
	if !IsKeyValid(key) {
		return false
	}
	if len(data) == 0 {
		return false
	}
	if len(signature) != ed25519.SignatureSize {
		return false
	}

	hash := sha256.Sum256(key[:])
	priv := ed25519.NewKeyFromSeed(hash[:])

	pub := priv.Public().(ed25519.PublicKey)
	return ed25519.Verify(pub, data, signature)
}
