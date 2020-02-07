// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compression

import (
	"context"
	"testing"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/security"
	"github.com/condensat/bank-core/security/utils"
)

func TestCompressMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, security.KeyPrivateKeySalt, utils.GenerateRandN(32))

	var data [32]byte
	message := bank.Message{
		Data: data[:],
	}
	compress := bank.Message{
		Data: data[:],
	}
	_ = CompressMessage(&compress, 5)

	encrypted := bank.Message{
		Data: data[:],
	}
	k := security.NewKey(ctx)
	_ = security.EncryptMessageFor(ctx, k, k.Public(ctx), &encrypted)

	type args struct {
		message *bank.Message
		level   int
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantCompress bool
	}{
		{"nil", args{nil, 5}, true, false},
		{"empty", args{new(bank.Message), 5}, true, false},
		{"encrypted", args{&encrypted, 5}, true, false},

		{"compress", args{&message, 5}, false, true},
		{"already_compress", args{&compress, 5}, false, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := CompressMessage(tt.args.message, tt.args.level); (err != nil) != tt.wantErr {
				t.Errorf("CompressMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.message == nil {
				return
			}

			if tt.args.message.IsCompressed() != tt.wantCompress {
				t.Errorf("CompressMessage() = %v, wantCompress %v", tt.args.message.IsCompressed(), tt.wantCompress)
			}
		})
	}
}

func TestDecompressMessage(t *testing.T) {
	t.Parallel()

	var data [32]byte
	message := bank.Message{
		Data: data[:],
	}
	compress := bank.Message{
		Data: data[:],
	}
	_ = CompressMessage(&compress, 5)

	type args struct {
		message *bank.Message
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantCompress bool
	}{
		{"nil", args{nil}, true, false},
		{"empty", args{new(bank.Message)}, true, false},
		{"compress", args{&compress}, false, false},
		{"not_compressed", args{&message}, false, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := DecompressMessage(tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("DecompressMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.message == nil {
				return
			}

			if tt.args.message.IsCompressed() != tt.wantCompress {
				t.Errorf("DecompressMessage() = %v, wantCompress %v", tt.args.message.IsCompressed(), tt.wantCompress)
			}
		})
	}
}
