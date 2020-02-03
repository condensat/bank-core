// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package utils

func Xor(result, key, buffer []byte) {
	if len(key) == 0 {
		panic("Invalid key len for Xor")
	}
	if len(result) != len(buffer) {
		panic("Result and buffer must have the same size")
	}
	for i, b := range buffer {
		result[i] = b ^ key[i%len(key)]
	}
}
