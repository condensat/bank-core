// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"errors"

	"github.com/condensat/bank-core/security/secureid/internal"

	"crypto/aes"
	"crypto/cipher"
)

// encrypt return encrypted SecureID from BlockData
func encrypt(keys *Keys, message internal.BlockData) (SecureID, error) {
	if keys == nil {
		return SecureID{}, errors.New("Invalid keys")
	}

	// derived key to encrypt data
	gen := internal.CreateKeyGenerator(keys.hash, keys.keyInfo)
	kAES, err := gen.NextKey()
	if err != nil {
		return SecureID{}, errors.New("Failed to get kAES")
	}
	// derived key to generate hmac signature
	kCRC, err := gen.NextKey()
	if err != nil {
		return SecureID{}, errors.New("Failed to get kkCRCSign")
	}
	// derived key to generate hmac iv
	kIV, err := gen.NextKey()
	if err != nil {
		return SecureID{}, errors.New("Failed to get kIV")
	}

	// generate iv from message and kIV
	iv := internal.HmacBlock(keys.hash, kIV, message[:])

	// cleartext message signature
	checksum := internal.Checksum(keys.hash, kCRC, message[:])
	if len(checksum) < 4 {
		return SecureID{}, errors.New("Invalid checksum")
	}

	// store 32 bits checksum after IV 96 bits
	copy(iv[12:], checksum[:4])

	block, _ := aes.NewCipher(kAES)
	mode := cipher.NewCBCEncrypter(block, iv[:])
	var ciphertext internal.BlockData
	mode.CryptBlocks(ciphertext[:], message[:])

	return SecureID{
		Version: VersionFormat(keys.Version, keys.KeyID),
		Data:    ciphertext.String(),
		Check:   iv.String(),
	}, nil
}
