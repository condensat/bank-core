// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"crypto/sha512"
	"fmt"
	"strconv"
	"strings"
)

const (
	Version0 = ProtocolVersion("v0") // see genKeyInfoV0
	Version1 = ProtocolVersion("v1") // see genKeyInfoV1

	Version = ProtocolVersion(Version1) // Current ProtocolVersion
)

var (
	DefaultHash = sha512.New // DefaultHash is sha512
)

// VersionFormat contruct Version from ProtocolVersion and KeyID
// see SecureID Version field
func VersionFormat(version ProtocolVersion, keyID KeyID) string {
	return fmt.Sprintf("%s.%d", version, keyID)
}

// VersionParse return ProtocolVersion and KeyID from Version
// see SecureID Version field
func VersionParse(str string) (ProtocolVersion, KeyID) {
	var version = string(Version)
	var keyID KeyID
	tok := strings.Split(str, ".")
	if len(tok) > 0 {
		version = tok[0]
	}
	if len(tok) > 1 {
		num, err := strconv.ParseUint(tok[1], 10, 64)
		if err == nil {
			keyID = KeyID(num)
		}
	}

	if len(version) == 0 {
		version = string(Version)
	}
	return ProtocolVersion(version), keyID
}

// HashFromVersion return HashFunc from ProtocolVersion
func HashFromVersion(version ProtocolVersion) HashFunc {
	switch version {

	case Version0:
		fallthrough
	case Version1:
		return sha512.New

	default:
		return DefaultHash
	}
}
