// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package libsodium

import (
	"bytes"
	"errors"

	"crypto/ed25519"

	"github.com/condensat/bank-core/security/utils"

	"golang.org/x/crypto/curve25519"
)

const (
	Curve25519Size = 32
)

func ConvertSecretKey(secretKey []byte) ([Curve25519Size]byte, error) {
	var result [Curve25519Size]byte
	err := crypto_sign_ed25519_sk_to_curve25519(result[:], secretKey[:])
	if err != nil {
		utils.Memzero(result[:])
		return result, err
	}

	return result, nil
}

func ConvertPublicKey(publicKey []byte) ([Curve25519Size]byte, error) {
	var result [Curve25519Size]byte
	err := crypto_sign_ed25519_pk_to_curve25519(result[:], publicKey[:])
	if err != nil {
		utils.Memzero(result[:])
		return result, err
	}

	return result, nil
}

func VerifyKeys(publicKey ed25519.PublicKey, secretKey ed25519.PrivateKey) error {
	if !bytes.Equal(publicKey[:], secretKey[32:]) {
		return errors.New("Private key do not contains public key")
	}

	// verify converted public key
	privKey, err := ConvertSecretKey(secretKey[:])
	defer utils.Memzero(privKey[:])
	if err != nil {
		return err
	}

	pubKey, err := ConvertPublicKey(publicKey[:])
	defer utils.Memzero(pubKey[:])
	if err != nil {
		return err
	}

	var verif [Curve25519Size]byte
	defer utils.Memzero(verif[:])
	curve25519.ScalarBaseMult(&verif, &privKey)

	if !bytes.Equal(pubKey[:], verif[:]) {
		return errors.New("Public keys does not match")
	}

	return nil
}
