// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bank

import (
	"reflect"
	"testing"
)

func TestMessage_SetCompressed(t *testing.T) {
	t.Parallel()

	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     string
	}
	type args struct {
		compressed bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint
	}{
		{"default_compressed", fields{}, args{true}, 1 << flagCompressed},
		{"default_not_compressed", fields{}, args{false}, 0},
		{"unsed_compressed", fields{Flags: 1 << flagCompressed}, args{false}, 0},
		{"keepflag_compressed", fields{Flags: (1 << flagEncrypted) + 1<<flagCompressed}, args{false}, (1 << flagEncrypted)},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version: tt.fields.Version,
				From:    tt.fields.From,
				Data:    tt.fields.Data,
				Flags:   tt.fields.Flags,
				Error:   tt.fields.Error,
			}
			m.SetCompressed(tt.args.compressed)

			if m.Flags != tt.want {
				t.Errorf("Message.SetCompressed() = %v, want %v", m.Flags, tt.want)
			}
		})
	}
}

func TestMessage_SetEncrypted(t *testing.T) {
	t.Parallel()

	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     string
	}
	type args struct {
		compressed bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint
	}{
		{"default_encrypted", fields{}, args{true}, 1 << flagEncrypted},
		{"default_not_encrypted", fields{}, args{false}, 0},
		{"unsed_encrypted", fields{Flags: 1 << flagEncrypted}, args{false}, 0},
		{"keepflag_encrypted", fields{Flags: (1 << flagEncrypted) + 1<<flagCompressed}, args{false}, (1 << flagCompressed)},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version: tt.fields.Version,
				From:    tt.fields.From,
				Data:    tt.fields.Data,
				Flags:   tt.fields.Flags,
				Error:   tt.fields.Error,
			}
			m.SetEncrypted(tt.args.compressed)

			if m.Flags != tt.want {
				t.Errorf("Message.SetEncrypted() = %v, want %v", m.Flags, tt.want)
			}
		})
	}
}

func TestMessage_SetSigned(t *testing.T) {
	t.Parallel()

	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     string
	}
	type args struct {
		compressed bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint
	}{
		{"default_not_signed", fields{}, args{false}, 0},
		{"unsed_signed", fields{Flags: 1 << flagSigned}, args{false}, 0},
		{"default_signed", fields{}, args{true}, 1 << flagSigned},
		{"keepflag_signed", fields{Flags: (1 << flagSigned) + 1<<flagCompressed}, args{false}, (1 << flagCompressed)},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version: tt.fields.Version,
				From:    tt.fields.From,
				Data:    tt.fields.Data,
				Flags:   tt.fields.Flags,
				Error:   tt.fields.Error,
			}
			m.SetSigned(tt.args.compressed)

			if m.Flags != tt.want {
				t.Errorf("Message.SetSigned() = %v, want %v", m.Flags, tt.want)
			}
		})
	}
}

func TestMessage_IsCompressed(t *testing.T) {
	t.Parallel()

	var m Message
	m.SetCompressed(true)
	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, false},
		{"compressed", fields{Flags: 1 << flagCompressed}, true},
		{"encrypted", fields{Flags: 1 << flagEncrypted}, false},
		{"signed", fields{Flags: 1 << flagSigned}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version: tt.fields.Version,
				From:    tt.fields.From,
				Data:    tt.fields.Data,
				Flags:   tt.fields.Flags,
				Error:   tt.fields.Error,
			}
			if got := m.IsCompressed(); got != tt.want {
				t.Errorf("Message.IsCompressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_IsEncrypted(t *testing.T) {
	t.Parallel()

	var m Message
	m.SetCompressed(true)
	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, false},
		{"compressed", fields{Flags: 1 << flagCompressed}, false},
		{"encrypted", fields{Flags: 1 << flagEncrypted}, true},
		{"signed", fields{Flags: 1 << flagSigned}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version: tt.fields.Version,
				From:    tt.fields.From,
				Data:    tt.fields.Data,
				Flags:   tt.fields.Flags,
				Error:   tt.fields.Error,
			}
			if got := m.IsEncrypted(); got != tt.want {
				t.Errorf("Message.IsEncrypted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_IsSigned(t *testing.T) {
	t.Parallel()

	var m Message
	m.SetCompressed(true)
	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, false},
		{"compressed", fields{Flags: 1 << flagCompressed}, false},
		{"encrypted", fields{Flags: 1 << flagEncrypted}, false},
		{"signed", fields{Flags: 1 << flagSigned}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version: tt.fields.Version,
				From:    tt.fields.From,
				Data:    tt.fields.Data,
				Flags:   tt.fields.Flags,
				Error:   tt.fields.Error,
			}
			if got := m.IsSigned(); got != tt.want {
				t.Errorf("Message.IsSigned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_Encode(t *testing.T) {
	t.Parallel()

	type fields struct {
		Version   string
		From      string
		Data      []byte
		Signature string
		Flags     uint
		Error     string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{"Default", fields{}, 76, false},
		{"Encode", fields{"1.0", "from", nil, "", 42, ""}, 89, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Version: tt.fields.Version,
				From:    tt.fields.From,
				Data:    tt.fields.Data,
				Flags:   tt.fields.Flags,
				Error:   tt.fields.Error,
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
	t.Parallel()

	message := Message{"1.0", "from", nil, 42, ""}
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
		tt := tt // capture range variable
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
