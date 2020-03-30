// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"fmt"

	"github.com/condensat/bank-core/security/secureid/internal"
)

// genDefaultKeyInfo current default KeyInfo generator
var genDefaultKeyInfo = genKeyInfoV1

// genKeyInfo return versioned KeyInfo generator from ProtocolVersion and KeyID
// KeyID is append to info.Context to generate hkdf info
func genKeyInfo(info SecureInfo, version ProtocolVersion, keyID KeyID, hash HashFunc) internal.KeyInfo {
	if hash == nil {
		panic("Invalid hash")
	}

	switch version {

	case Version0:
		return genKeyInfoV0(info, keyID, hash)
	case Version1:
		return genKeyInfoV1(info, keyID, hash)

	default:
		return genDefaultKeyInfo(info, keyID, hash)
	}
}

// genKeyInfoV0 derive secret and salt from seed and create context from info.Context and KeyID
func genKeyInfoV0(info SecureInfo, keyID KeyID, hash HashFunc) internal.KeyInfo {
	if len(info.Seed) < internal.BlockSize {
		panic("Invalid seed")
	}
	// compute secret key from seed
	h := hash()
	_, _ = h.Write(info.Seed)
	secret := h.Sum(nil)

	// compute salt key from seed and secret
	_, _ = h.Write(secret)
	salt := h.Sum(nil)

	// compute context info from context and keyID
	context := fmt.Sprintf("%s:%d", info.Context, keyID)
	_, _ = h.Write([]byte(context))
	contextInfo := h.Sum(nil)

	return internal.KeyInfo{
		Secret: secret,
		Salt:   salt,
		Info:   contextInfo,
	}
}

// genKeyInfoV1 reverse KeyInfo returned by genKeyInfoV0
func genKeyInfoV1(info SecureInfo, keyID KeyID, hash HashFunc) internal.KeyInfo {
	keyInfo := genKeyInfoV0(info, keyID, hash)

	// reverse all KeyInfo bytes
	keyInfo.Reverse()

	return keyInfo
}
