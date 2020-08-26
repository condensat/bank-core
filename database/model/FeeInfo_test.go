// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"
)

func TestFeeInfo_IsValid(t *testing.T) {
	const currency = "CURR"
	const minimumFee = 1.0

	type fields struct {
		Currency CurrencyName
		Minimum  Float
		Rate     Float
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, false},
		{"invalid_currency", fields{"", minimumFee, DefaultFeeRate}, false},
		{"invalid_minimum", fields{currency, -minimumFee, DefaultFeeRate}, false},
		{"invalid_rate", fields{currency, minimumFee, -DefaultFeeRate}, false},

		{"zero", fields{currency, 0.0, 0.0}, true},
		{"valid", fields{currency, minimumFee, DefaultFeeRate}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FeeInfo{
				Currency: tt.fields.Currency,
				Minimum:  tt.fields.Minimum,
				Rate:     tt.fields.Rate,
			}
			if got := p.IsValid(); got != tt.want {
				t.Errorf("FeeInfo.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeInfo_Compute(t *testing.T) {
	const currency = "CURR"
	const minimumFee = 1.0

	type fields struct {
		Currency CurrencyName
		Minimum  Float
		Rate     Float
	}
	type args struct {
		amount Float
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Float
	}{
		{"default", fields{}, args{}, 0.0},
		{"zero", fields{}, args{42.0}, 0.0},

		{"min_less", fields{currency, minimumFee, DefaultFeeRate}, args{999.0}, 1.0},
		{"minimum", fields{currency, minimumFee, DefaultFeeRate}, args{1000.0}, 1.0},
		{"min_more", fields{currency, minimumFee, DefaultFeeRate}, args{1001.0}, 1.001},

		{"small", fields{currency, minimumFee, DefaultFeeRate}, args{500.0}, 1.0},
		{"percent", fields{currency, minimumFee, DefaultFeeRate}, args{1337.0}, 1.337},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FeeInfo{
				Currency: tt.fields.Currency,
				Minimum:  tt.fields.Minimum,
				Rate:     tt.fields.Rate,
			}
			if got := p.Compute(tt.args.amount); got != tt.want {
				t.Errorf("FeeInfo.Compute() = %v, want %v", got, tt.want)
			}
		})
	}
}
