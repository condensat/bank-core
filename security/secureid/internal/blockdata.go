// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package internal

import (
	"bytes"
	"errors"
	"hash"

	"crypto/aes"
	"crypto/hmac"

	"encoding/binary"
)

const (
	BlockSize = aes.BlockSize
)

// BlockData used for aes operations
type BlockData [BlockSize]byte

func (p BlockData) String() string {
	return ToUUID(p[:]).String()
}

// Store value into current BlockData
func (p *BlockData) Store(value interface{}) error {
	switch data := value.(type) {
	case uint64:
		return p.StoreUInt64(data)

	case []byte:
		return p.StoreBytes(data)

	default:
		return errors.New("Unknown data type")
	}
}

// StoreBytes write data into current BlockData
func (p *BlockData) StoreBytes(data []byte) error {
	if len(data) > BlockSize {
		return errors.New("Data too long")
	}
	copy(p[:], data[:])
	return nil
}

// StoreUInt64 write uint64 value into current BlockData
func (p *BlockData) StoreUInt64(value uint64) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		return err
	}

	if buf.Len() < BlockSize {
		_, err = buf.Write(make([]byte, BlockSize-buf.Len()))
		if err != nil {
			return err
		}
	}

	if buf.Len() != BlockSize {
		return err
	}

	copy(p[:], buf.Bytes())
	return nil
}

// UInt64 retrieve uint64 value from current BlockData
func (p *BlockData) UInt64() uint64 {
	buf := bytes.NewBuffer(p[:])
	var value uint64
	err := binary.Read(buf, binary.LittleEndian, &value)
	if err != nil {
		return 0
	}

	return value
}

// HmacBlock compute hmac from hash and key.
// hmac is stored and returned into BlockData
func HmacBlock(hash func() hash.Hash, key, data []byte) BlockData {
	hm := hmac.New(hash, key)
	_, _ = hm.Write(data)

	var result BlockData
	_ = result.StoreBytes(hm.Sum(nil)[:BlockSize])

	return result
}
