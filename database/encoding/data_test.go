// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"
)

type CustomData struct {
	Foo string `json:"foo,omitempty"`
}

func TestEncodeData(t *testing.T) {
	t.Parallel()

	const emptyData = Data("")
	var invalidJson = func() {}

	type args struct {
		instance DataInterface
	}
	tests := []struct {
		name    string
		args    args
		want    Data
		wantErr bool
	}{
		{"default", args{}, emptyData, false},
		{"not_serializable", args{&invalidJson}, emptyData, true},

		{"data_empty", args{&CustomData{}}, Data("{}"), false},
		{"data_value", args{&CustomData{"bar"}}, Data(`{"foo":"bar"}`), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeData(tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EncodeData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeData(t *testing.T) {
	t.Parallel()

	const invalidData = Data("{")
	const emptyData = Data("")
	validData, _ := EncodeData(&CustomData{"bar"})

	type args struct {
		instance DataInterface
		data     Data
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"default", args{}, false},
		{"invalid", args{&CustomData{}, invalidData}, true},

		{"empty_data", args{&CustomData{}, emptyData}, false},
		{"data", args{&CustomData{"bar"}, validData}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := DecodeData(tt.args.instance, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("DecodeData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
