// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"errors"

	"crypto/aes"
	"crypto/cipher"

	"github.com/condensat/bank-core/security/secureid/internal"
)

// decrypt return decrypted BlockData from SecureID
func decrypt(keys *Keys, secureID SecureID) (internal.BlockData, error) {
	version, keyID := VersionParse(secureID.Version)
	if keys == nil {
		return internal.BlockData{}, errors.New("Invalid keys")
	}
	if keys.Version != version {
		return internal.BlockData{}, errors.New("Incompatible version")
	}
	if keys.KeyID != keyID {
		return internal.BlockData{}, errors.New("Incompatible keyID")
	}

	// derived key to decrypt data
	gen := internal.CreateKeyGenerator(keys.hash, keys.keyInfo)
	kAES, err := gen.NextKey()
	if err != nil {
		return internal.BlockData{}, errors.New("Failed to get kAES")
	}
	// derived key to generate hmac signature
	kCRC, err := gen.NextKey()
	if err != nil {
		return internal.BlockData{}, errors.New("Failed to get kCRC")
	}

	ciphertext := internal.FromUUID(secureID.Data)
	iv := internal.FromUUID(secureID.Check)

	block, _ := aes.NewCipher(kAES)
	mode := cipher.NewCBCDecrypter(block, iv)
	var message internal.BlockData
	mode.CryptBlocks(message[:], ciphertext)

	// extract 32 bits checksum from IV
	checksum := iv[12:]

	if !internal.Verify(keys.hash, kCRC, message[:], checksum) {
		message = internal.BlockData{}
		return message, errors.New("Wrong message signature")
	}

	return message, nil
}
