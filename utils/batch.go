// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package utils

func BatchString(batchSize int, strings ...string) [][]string {
	batches := make([][]string, 0, (len(strings)+batchSize-1)/batchSize)

	for batchSize < len(strings) {
		strings, batches = strings[batchSize:], append(batches, strings[0:batchSize:batchSize])
	}
	batches = append(batches, strings)

	return batches[:]
}
