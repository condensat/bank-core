// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bank

import (
	"bytes"
	"encoding/gob"
)

const (
	cstCurrentVersion = "1.0"
)

// Message used for all communication between components
type Message struct {
	Version   string // Version for compatibility
	From      string `json:",omitempty"` // From is the public key of sender
	Data      []byte `json:",omitempty"` // Data payload
	Signature string `json:",omitempty"` // Signature of data with sender private key
	Flags     uint   `json:",omitempty"` // Flags for Compressed, Encrypted, Signed
	Error     error  `json:",omitempty"` // Error in message processing
}

func NewMessage() *Message {
	return &Message{
		Version: cstCurrentVersion,
	}
}

// Encode return bytes from Message. Encoded with gob
func (m *Message) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)

	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode return Message from bytes. Decoded with gob
func (m *Message) Decode(data []byte) error {
	buffer := bytes.NewReader(data)
	dec := gob.NewDecoder(buffer)

	err := dec.Decode(m)
	if err != nil {
		return err
	}
	return nil
}
