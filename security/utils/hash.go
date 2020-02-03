// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package utils

import (
	"crypto/sha512"
)

func HashString(str string) []byte {
	if len(str) == 0 {
		panic("Invalid string to hash")
	}
	return HashBytes([]byte(str))
}

func HashBytes(buff []byte) []byte {
	if len(buff) == 0 {
		panic("Invalid buff to hash")
	}
	return HashBuffers(buff)
}

func HashBuffers(buffers ...[]byte) []byte {
	if len(buffers) == 0 {
		panic("Invalid buffers to hash")
	}
	h := sha512.New()
	defer h.Reset()
	for _, buff := range buffers {
		_, _ = h.Write(buff)
	}
	return h.Sum(nil)
}
