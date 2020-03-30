// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"hash"
)

// ProtocolVersion type for protocol versioning
type ProtocolVersion string

// KeyID type for keys rotation
type KeyID uint

// Seed type used to init secrets for Keys
type Seed []byte

// Value type used as payload of transmited data
type Value uint64

// HashFunc is used for extenal hash algorithm
// default is SHA512
type HashFunc func() hash.Hash

// SecureInfo struct used for SecureID function and Keys initialisation
type SecureInfo struct {
	Seed    Seed   // seed used for keys derivation
	Context string // context used for hkdf
}
