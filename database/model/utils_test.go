// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "testing"

func TestToFixedFloat(t *testing.T) {
	type args struct {
		value Float
	}
	tests := []struct {
		name string
		args args
		want Float
	}{
		{"zero", args{0.0}, 0.0},
		{"piMax", args{3.14159265358979}, 3.14159265359},
		{"piOverflow", args{3.14159265358979}, 3.14159265359},
		{"roundFloor", args{0.1234567890121}, 0.123456789012},
		{"roundCeil", args{0.1234567890129}, 0.123456789013},

		{"roundBig", args{2.1e+11}, 2.1e+11},
		{"roundSmall", args{2.1e-11}, 2.1e-11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToFixedFloat(tt.args.value); got != tt.want {
				t.Errorf("ToFixedFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}
