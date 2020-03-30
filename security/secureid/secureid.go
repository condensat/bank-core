// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"errors"
)

// SecureID struct
type SecureID struct {
	Version string `json:"ver"`

	Data  string `json:"id"`
	Check string `json:"sig"`
}

// SecureIDFromValue convert Value to SecureID using SecureInfo, ProtocolVersion and KeyID
func SecureIDFromValue(info SecureInfo, version ProtocolVersion, keyID KeyID, value Value) (SecureID, error) {
	keys := NewKeys(info, version, keyID)
	if keys == nil {
		return SecureID{}, errors.New("Failed to create keys")
	}
	return keys.SecureIDFromValue(value)
}

// Value convert SecureID to Value SecureInfo
// Version and KeyID are retrived from secureID.Version field
func (p *SecureID) Value(info SecureInfo, secureID SecureID) (Value, error) {
	version, keyID := VersionParse(p.Version)
	keys := NewKeys(info, version, keyID)
	if keys == nil {
		return 0, errors.New("Failed to create keys")
	}

	return keys.ValueFromSecureID(secureID)
}
