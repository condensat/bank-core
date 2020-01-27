// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"context"

	"crypto/subtle"

	"github.com/condensat/bank-core/logger"
	"golang.org/x/crypto/argon2"
)

// SaltedHash return argon2 key from salt and input
func SaltedHash(ctx context.Context, password, salt []byte) []byte {
	if len(password) == 0 || len(salt) < 8 {
		logger.Logger(ctx).
			Panic("Invalid Input")
	}

	return argon2.Key(password, salt, 1, 32<<10, 4, 32)
}

// SaltedHashVerify check if key correspond to argon2 hash from salt and input
func SaltedHashVerify(ctx context.Context, password, salt []byte, key []byte) bool {
	if len(password) == 0 || len(salt) < 8 {
		return false
	}

	// compute argon hash
	dk := argon2.Key(password, salt, 1, 32<<10, 4, 32)

	// check if length match
	if subtle.ConstantTimeEq(int32(len(dk)), int32(len(key))) == 0 {
		return false
	}

	// Compare keys content
	return subtle.ConstantTimeCompare(dk, key) == 1
}
