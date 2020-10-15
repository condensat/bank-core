// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"github.com/condensat/bank-core/compression"
)

func CompressMessage(message *Message, level int) error {
	if message == nil {
		return ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return ErrNoData
	}

	if message.IsCompressed() {
		// NOOP
		return nil
	}

	if message.IsEncrypted() {
		return ErrOperationNotPermited
	}

	data, err := compression.Compress(message.Data, level)
	if err != nil {
		return err
	}
	message.Data = data
	message.SetCompressed(true)

	return nil
}

func DecompressMessage(message *Message) error {
	if message == nil {
		return ErrInvalidMessage
	}
	if len(message.Data) == 0 {
		return ErrNoData
	}

	if !message.IsCompressed() {
		// NOOP
		return nil
	}

	data, err := compression.Decompress(message.Data)
	if err != nil {
		return err
	}
	message.Data = data
	message.SetCompressed(false)

	return nil
}
