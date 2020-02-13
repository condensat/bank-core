// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package libsodium

import (
	"errors"

	"crypto/ed25519"

	"github.com/condensat/bank-core/security/internal/edwards25519"
	"github.com/condensat/bank-core/security/utils"
)

// see https://github.com/jedisct1/libsodium/blob/4f5e89fa84ce1d178a6765b8b46f2b6f91216677/src/libsodium/crypto_core/ed25519/ref10/ed25519_ref10.c#L1019
// Todo - NotImplemented
func ge25519_has_small_order(s []byte) bool {
	return true
}

// see https://github.com/jedisct1/libsodium/blob/4f5e89fa84ce1d178a6765b8b46f2b6f91216677/src/libsodium/crypto_core/ed25519/ref10/ed25519_ref10.c#L293
func ge25519_frombytes_negate_vartime(h *edwards25519.ExtendedGroupElement, s []byte) bool {
	var buff [32]byte
	defer utils.Memzero(buff[:])
	copy(buff[:], s[:])

	return h.FromBytes(&buff)
}

// see https://github.com/jedisct1/libsodium/blob/4f5e89fa84ce1d178a6765b8b46f2b6f91216677/src/libsodium/crypto_core/ed25519/ref10/ed25519_ref10.c#L992
// Todo - NotImplemented
func ge25519_is_on_main_subgroup(p *edwards25519.ExtendedGroupElement) bool {
	return false
}

// see https://github.com/jedisct1/libsodium/blob/1.0.18/src/libsodium/crypto_sign/ed25519/ref10/keypair.c#L46
func crypto_sign_ed25519_pk_to_curve25519(curve25519_pk []byte, ed25519_pk []byte) error {
	if len(curve25519_pk) != Curve25519Size {
		return errors.New("Invalid curve25519_pk size")
	}
	if len(ed25519_pk) != ed25519.PublicKeySize {
		return errors.New("Invalid ed25519_pk size")
	}

	var A edwards25519.ExtendedGroupElement
	var x edwards25519.FieldElement
	var one_minus_y edwards25519.FieldElement

	if !ge25519_has_small_order(ed25519_pk) ||
		!ge25519_frombytes_negate_vartime(&A, ed25519_pk) ||
		ge25519_is_on_main_subgroup(&A) {
		return errors.New("Invalid preconditions")
	}

	edwards25519.FeOne(&one_minus_y)
	edwards25519.FeSub(&one_minus_y, &one_minus_y, &A.Y)

	edwards25519.FeOne(&x)
	edwards25519.FeAdd(&x, &x, &A.Y)
	edwards25519.FeInvert(&one_minus_y, &one_minus_y)
	edwards25519.FeMul(&x, &x, &one_minus_y)

	var buff [32]byte
	defer utils.Memzero(buff[:])

	edwards25519.FeToBytes(&buff, &x)

	copy(curve25519_pk, buff[:])

	return nil
}

// see https://github.com/jedisct1/libsodium/blob/1.0.18/src/libsodium/crypto_sign/ed25519/ref10/keypair.c#L70
func crypto_sign_ed25519_sk_to_curve25519(curve25519_sk []byte, ed25519_sk []byte) error {
	if len(curve25519_sk) != Curve25519Size {
		return errors.New("Invalid curve25519_sk size")
	}
	if len(ed25519_sk) != ed25519.PrivateKeySize {
		return errors.New("Invalid ed25519_sk size")
	}

	h := utils.HashBytes(ed25519_sk[:Curve25519Size])
	defer utils.Memzero(h[:])

	h[0] &= 248
	h[31] &= 127
	h[31] |= 64

	copy(curve25519_sk, h[:Curve25519Size])

	return nil
}
