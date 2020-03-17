// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compression

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"

	"github.com/condensat/bank-core"
)

var (
	ErrCompress   = errors.New("Compress Error")
	ErrDecompress = errors.New("Decompress Error")
)

func clamp(count, min, max int) int {
	if count < min {
		return min
	} else if count > max {
		return max
	} else {
		return count
	}
}

func Compress(data []byte, level int) ([]byte, error) {
	if len(data) == 0 {
		return nil, bank.ErrNoData
	}
	level = clamp(level, 0, 9)

	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, level)
	if err != nil {
		return nil, ErrCompress
	}
	l, err := w.Write(data[:])
	w.Close()
	if err != nil {
		return nil, ErrCompress
	}
	if l != len(data) {
		return nil, ErrCompress
	}

	return b.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, bank.ErrNoData
	}

	b := bytes.NewBuffer(data[:])
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, ErrDecompress
	}
	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, ErrDecompress
	}

	return data, nil
}
