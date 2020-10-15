// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"bytes"
	"encoding/gob"
)

type BankObject interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}

// encode return bytes from BankObject. Encoded with gob
func EncodeObject(object BankObject) ([]byte, error) {
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)

	err := enc.Encode(object)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode store BankObject from bytes. Decoded with gob
func DecodeObject(data []byte, object BankObject) error {
	buffer := bytes.NewReader(data)
	dec := gob.NewDecoder(buffer)

	err := dec.Decode(object)
	if err != nil {
		return err
	}
	return nil
}

func ToMessage(from string, object BankObject) *Message {
	data, err := object.Encode()
	if err != nil {
		return nil
	}
	return &Message{
		Version: CurrentVersion,
		From:    from,
		Data:    data,
	}
}

func FromMessage(message *Message, object BankObject) error {
	return object.Decode(message.Data)
}
