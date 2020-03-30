// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"github.com/condensat/bank-core/security/secureid/internal"
)

// Keys struct holds state for encrypt/decrypt SecureID
type Keys struct {
	Version ProtocolVersion
	KeyID   KeyID

	// private fields
	hash    HashFunc
	keyInfo internal.KeyInfo
}

// DefaultKeys return Keys from current ProtocolVersion and KeyID, with SecureInfo
func DefaultKeys(info SecureInfo, keyID KeyID) *Keys {
	return NewKeys(info, Version, keyID)
}

// NewKeys return Key from SecureInfo, ProtocolVersion and KeyID
func NewKeys(info SecureInfo, version ProtocolVersion, keyID KeyID) *Keys {
	return NewKeysWithHashFunction(info, version, keyID, nil)
}

// NewKeysWithHashFunction return Key from SecureInfo, ProtocolVersion, KeyID and HashFunc
// ProtocolVersion hash is used if nil
func NewKeysWithHashFunction(info SecureInfo, version ProtocolVersion, keyID KeyID, hash HashFunc) *Keys {
	if hash == nil {
		hash = HashFromVersion(version)
	}
	keyInfo := genKeyInfo(info, version, keyID, hash)
	if !keyInfo.Valid() {
		return nil
	}

	return &Keys{
		Version: version,
		KeyID:   keyID,
		hash:    hash,
		keyInfo: keyInfo,
	}
}

// SecureIDFromValue return SecureID from Value using Keys state
func (p *Keys) SecureIDFromValue(value Value) (SecureID, error) {
	var data internal.BlockData
	err := data.StoreUInt64(uint64(value))
	if err != nil {
		return SecureID{}, err
	}
	return encrypt(p, data)
}

// ValueFromSecureID return Value from SecureID using Keys state
func (p *Keys) ValueFromSecureID(secureID SecureID) (Value, error) {
	data, err := decrypt(p, secureID)
	if err != nil {
		return 0, err
	}

	return Value(data.UInt64()), nil
}
