// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"errors"

	"github.com/condensat/bank-core/security/utils"

	"golang.org/x/crypto/nacl/box"
)

func EncryptFor(from EncryptionPrivateKey, to EncryptionPublicKey, message []byte) ([]byte, error) {
	pubKey := [32]byte(to)
	privKey := [32]byte(from)
	defer utils.Memzero(pubKey[:])
	defer utils.Memzero(privKey[:])

	nonce, err := GenerateNonce()
	defer utils.Memzero(nonce[:])
	if err != nil {
		return nil, err
	}

	data := box.Seal(nonce[:], message[:], &nonce, &pubKey, &privKey)
	if len(data) == 0 {
		return nil, errors.New("Box seal failed")
	}

	return data, nil
}

func DecryptFrom(from EncryptionPublicKey, to EncryptionPrivateKey, data []byte) ([]byte, error) {
	if len(data) <= NonceSize {
		return nil, errors.New("Invalid data")
	}
	pubKey := [32]byte(from)
	privKey := [32]byte(to)
	var nonce [NonceSize]byte
	defer utils.Memzero(pubKey[:])
	defer utils.Memzero(privKey[:])
	defer utils.Memzero(nonce[:])

	copy(nonce[:], data[:NonceSize])
	data, ok := box.Open(nil, data[NonceSize:], &nonce, &pubKey, &privKey)
	if !ok {
		return nil, errors.New("Fail to open box")
	}

	return data, nil
}
