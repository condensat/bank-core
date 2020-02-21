// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bank

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/emef/bitfield"
)

type MessageFlag uint32

const (
	CurrentVersion = "1.0"

	flagCompressed MessageFlag = 0
	flagEncrypted  MessageFlag = 1
	flagSigned     MessageFlag = 2
)

var (
	ErrInvalidMessage = errors.New("Invalid Message")
	ErrNoData         = errors.New("No Data")
)

// Message used for all communication between components
type Message struct {
	Version string // Version for compatibility
	From    string `json:",omitempty"` // From is the public key of sender
	Data    []byte `json:",omitempty"` // Data payload
	Flags   uint   `json:",omitempty"` // Flags for Compressed, Encrypted, Signed
	Error   string `json:",omitempty"` // Error in message processing
}

func NewMessage() *Message {
	return &Message{
		Version: CurrentVersion,
	}
}

func (m *Message) SetCompressed(compressed bool) {
	m.setFlag(flagCompressed, compressed)
}

func (m *Message) SetEncrypted(encrypted bool) {
	m.setFlag(flagEncrypted, encrypted)
}

func (m *Message) SetSigned(signed bool) {
	m.setFlag(flagSigned, signed)
}

func (m *Message) IsCompressed() bool {
	return m.getFlag(flagCompressed)
}

func (m *Message) IsEncrypted() bool {
	return m.getFlag(flagEncrypted)
}

func (m *Message) IsSigned() bool {
	return m.getFlag(flagSigned)
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

func (m *Message) getFlag(flag MessageFlag) bool {
	b := bitfield.NewFromUint32(uint32(m.Flags))
	f := uint32(flag)
	return b.Test(f)
}

func (m *Message) setFlag(flag MessageFlag, enable bool) {
	b := bitfield.NewFromUint32(uint32(m.Flags))
	f := uint32(flag)

	if enable {
		b.Set(f)
	} else {
		b.Clear(f)
	}

	m.Flags = uint(b.ToUint32Safe())
}
