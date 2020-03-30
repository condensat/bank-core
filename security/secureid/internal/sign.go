// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package internal

import (
	"hash"

	"crypto/subtle"
)

// Checksum return data check sum
// hmac is computed from hashed data
func Checksum(hash func() hash.Hash, key, data []byte) []byte {
	h := hash()
	_, _ = h.Write(data)
	data = h.Sum(nil)

	hm := HmacBlock(hash, key, data)
	return hm[:]
}

// Verify return true if checksum match computed checksum from data
func Verify(hash func() hash.Hash, key, data, checksum []byte) bool {
	sign := Checksum(hash, key, data)
	sign = sign[:len(checksum)]
	return subtle.ConstantTimeCompare(sign, checksum) == 1
}
