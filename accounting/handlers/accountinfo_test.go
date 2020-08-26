// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"testing"
)

func Test_convertAssetAmount(t *testing.T) {
	type args struct {
		amount          float64
		tickerPrecision int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		// https://en.bitcoin.it/wiki/Units
		{"BTC", args{1.0, 8}, 1.0},          // 1 BTC = 1 BTC
		{"mBTC", args{0.001, 5}, 1},         // 1 mBTC 0.001 BTC
		{"bits", args{0.000001, 2}, 1},      // 1 bits 0.00001 BTC
		{"finney", args{0.0000001, 1}, 1},   // 1 finney = 0.0000001 BTC
		{"Satoshi", args{0.00000001, 0}, 1}, // 1 satoshi = 0.00000001 BTC
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertAssetAmount(tt.args.amount, tt.args.tickerPrecision); got != tt.want {
				t.Errorf("convertAssetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertAssetAmountToBitcoin(t *testing.T) {
	type args struct {
		amount          float64
		tickerPrecision int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		// https://en.bitcoin.it/wiki/Units
		{"BTC", args{1.0, 8}, 1.0},          // 1 BTC = 1 BTC
		{"mBTC", args{1, 5}, 0.001},         // 1 mBTC 0.001 BTC
		{"bits", args{1, 2}, 0.000001},      // 1 bits 0.00001 BTC
		{"finney", args{1, 1}, 0.0000001},   // 1 finney = 0.0000001 BTC
		{"Satoshi", args{1, 0}, 0.00000001}, // 1 satoshi = 0.00000001 BTC
		{"Asset", args{500000.0, 0}, 0.005}, // 💩00K  == 0.00💩BTC
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertAssetAmountToBitcoin(tt.args.amount, tt.args.tickerPrecision); got != tt.want {
				t.Errorf("convertAssetAmountToBitcoin() = %v, want %v", got, tt.want)
			}
		})
	}
}
