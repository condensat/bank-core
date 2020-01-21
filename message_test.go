// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bank

import (
	"reflect"
	"testing"
)

func TestMessage_Encode_(t *testing.T) {
	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     error
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{"Default", fields{}, 90, false},
		{"Encode", fields{"1.0", "from", nil, "", 42, nil}, 103, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version:   tt.fields.Version,
				From:      tt.fields.From,
				Data:      tt.fields.Data,
				Signature: tt.fields.Signature,
				Flags:     tt.fields.Flags,
				Error:     tt.fields.Error,
			}
			got, err := m.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Message.Encode() return %d bytes, want %d", len(got), tt.want)
			}
		})
	}
}

func TestMessage_Decode(t *testing.T) {
	message := Message{"1.0", "from", nil, "", 42, nil}
	data, _ := message.Encode()

	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     error
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Message
		wantErr bool
	}{
		{"Nil", fields{}, args{nil}, &Message{}, true},
		{"Decode", fields{}, args{data}, &message, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{}
			if err := m.Decode(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Message.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m, tt.want) {
				t.Errorf("Message.Encode() = %v, want %v", m, tt.want)
			}

		})
	}
}
