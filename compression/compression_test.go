// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compression

import (
	"bytes"
	"testing"
)

func TestCompress(t *testing.T) {
	t.Parallel()

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	type args struct {
		data  []byte
		level int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"nil", args{nil, 5}, 0, true},
		{"zero", args{zero[:], 5}, 0, true},

		{"level", args{data[:], -1}, 60, false},
		{"level2", args{data[:], 10}, 27, false},
		{"compress", args{data[:], 5}, 27, false},
		{"compress1", args{data1[:], 5}, 27, false},
		{"compress2", args{data2[:], 5}, 27, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compress(tt.args.data, tt.args.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Compress() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestDecompress(t *testing.T) {
	t.Parallel()

	var zero [0]byte
	var data [32]byte
	var data1 [64]byte
	var data2 [128]byte

	copmress, _ := Compress(data[:], 5)
	copmress1, _ := Compress(data1[:], 5)
	copmress2, _ := Compress(data2[:], 5)

	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"nil", args{nil}, nil, true},
		{"zero", args{zero[:]}, nil, true},
		{"compress", args{copmress[:]}, data[:], false},
		{"compress1", args{copmress1[:]}, data1[:], false},
		{"compress2", args{copmress2[:]}, data2[:], false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decompress(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Decompress() = %v, want %v", len(got), len(tt.want))
			}
		})
	}
}
