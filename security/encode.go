// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"errors"

	"github.com/condensat/bank-core/security/utils"

	"github.com/shengdoushi/base58"
)

var (
	DefaultAlphabet = base58.BitcoinAlphabet
)

// encoding/decoding helpers

func convert(from, to *base58.Alphabet, encoded string) string {
	data, _ := base58.Decode(encoded, from)
	return base58.Encode(data, to)
}

func encodeKey(key []byte) string {
	return base58.Encode(key[:], DefaultAlphabet)
}

func decodeKey(encoded string, result []byte) error {
	key, err := base58.Decode(encoded, DefaultAlphabet)
	defer utils.Memzero(key[:])
	if err != nil {
		return err
	}
	if len(key) != len(result) {
		return errors.New("Invalid key size")
	}

	copy(result[:], key)
	return nil
}

// Seed Key

func EncodeSeedKey(key SeedKey) string {
	defer utils.Memzero(key[:])
	return encodeKey(key[:])
}

func DecodeSeedKey(encoded string) (SeedKey, error) {
	var result SeedKey
	err := decodeKey(encoded, result[:])
	return result, err
}

// Signature Key

func EncodeSignatureKey(key SignaturePublicKey) string {
	defer utils.Memzero(key[:])
	return encodeKey(key[:])
}

func DecodeSignatureKey(encoded string) (SignaturePublicKey, error) {
	var result SignaturePublicKey
	err := decodeKey(encoded, result[:])
	return result, err
}

// Authentication Key

func EncodeAuthenticationKey(key AuthenticationKey) string {
	defer utils.Memzero(key[:])
	return encodeKey(key[:])
}

func DecodeAuthenticationKey(encoded string) (AuthenticationKey, error) {
	var result AuthenticationKey
	err := decodeKey(encoded, result[:])
	return result, err
}

func EncodeAuthenticationDigest(key AuthenticationDigest) string {
	defer utils.Memzero(key[:])
	return encodeKey(key[:])
}

func DecodeAuthenticationDigest(encoded string) (AuthenticationDigest, error) {
	var result AuthenticationDigest
	err := decodeKey(encoded, result[:])
	return result, err
}

// Public Key

func EncodePublicKey(key EncryptionPublicKey) string {
	defer utils.Memzero(key[:])
	return encodeKey(key[:])
}

func DecodePublicKey(encoded string) (EncryptionPublicKey, error) {
	var result EncryptionPublicKey
	err := decodeKey(encoded, result[:])
	return result, err
}
