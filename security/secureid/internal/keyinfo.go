// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package internal

// KeyInfo struct used for hkdf key derivation
type KeyInfo struct {
	Secret []byte // master secret key
	Salt   []byte // see crypto/hkdf
	Info   []byte // see crypto/hkdf
}

// Valid return true if KeyInfo is valid for crypto/hkdf
func (p *KeyInfo) Valid() bool {
	return len(p.Secret) > 0 && len(p.Salt) > 0 && len(p.Info) > 0
}

// Reverse KeyInfo data
// see secureid.Version0
func (p *KeyInfo) Reverse() {
	ReverseBytes(p.Secret)
	ReverseBytes(p.Salt)
	ReverseBytes(p.Info)
}
