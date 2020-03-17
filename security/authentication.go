// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"github.com/condensat/bank-core/security/utils"

	"golang.org/x/crypto/nacl/auth"
)

func AuthenticateMessage(authenticateKey AuthenticationKey, message []byte) AuthenticationDigest {
	defer utils.Memzero(authenticateKey[:])

	key := [AuthenticationKeySize]byte(authenticateKey)
	defer utils.Memzero(key[:])

	digest := auth.Sum(message, &key)
	defer utils.Memzero(digest[:])

	var auth AuthenticationDigest
	copy(auth[:], digest[:])
	return auth
}

func VerifyMessageAuthentication(authenticateKey AuthenticationKey, digest AuthenticationDigest, message []byte) bool {
	defer utils.Memzero(authenticateKey[:])

	key := [AuthenticationKeySize]byte(authenticateKey)
	defer utils.Memzero(key[:])

	return auth.Verify(digest[:], message, &key)
}
