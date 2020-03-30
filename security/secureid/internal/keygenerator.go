// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package internal

import (
	"hash"
	"io"

	"golang.org/x/crypto/hkdf"
)

// KeyGenerator for key dérivation
type KeyGenerator interface {
	NextKey() ([]byte, error)
}

// CreateKeyGenerator factory
func CreateKeyGenerator(hash func() hash.Hash, keyInfo KeyInfo) KeyGenerator {
	return &HkdfKeyGenerator{
		hkdf: hkdf.New(hash, keyInfo.Secret, keyInfo.Salt, keyInfo.Info),
	}
}

// HkdfKeyGenerator for hkdf keys derivation
type HkdfKeyGenerator struct {
	hkdf io.Reader
}

// NextKey return next hkdf key with BlockSize length
func (p *HkdfKeyGenerator) NextKey() ([]byte, error) {
	key := make([]byte, BlockSize)
	if _, err := io.ReadFull(p.hkdf, key); err != nil {
		return nil, err
	}

	return key, nil
}
