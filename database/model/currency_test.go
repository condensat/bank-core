// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
)

func newInt(value Int) ZeroInt {
	return &value
}

func TestNewCurrency(t *testing.T) {
	t.Parallel()

	type args struct {
		name      CurrencyName
		available Int
		crypto    Int
		precision Int
	}
	tests := []struct {
		name string
		args args
		want Currency
	}{
		{"Default", args{"", 0, 0, 0}, Currency{}},
		{"InvalidCurrency", args{"", 1, 1, 2}, Currency{}},
		{"DefaultAvailable", args{"BTC", -1, 1, 2}, Currency{"BTC", newInt(0), newInt(1), newInt(2)}},
		{"DefaultCrypto", args{"BTC", 1, -1, 2}, Currency{"BTC", newInt(1), newInt(0), newInt(2)}},
		{"DefaultPrecision", args{"BTC", 1, 1, -2}, Currency{"BTC", newInt(1), newInt(1), newInt(2)}},

		{"Valid", args{"BTC", 0, 0, 0}, Currency{"BTC", newInt(0), newInt(0), newInt(0)}},
		{"ValidAvailable", args{"BTC", 1, 0, 0}, Currency{"BTC", newInt(1), newInt(0), newInt(0)}},
		{"ValidCrypto", args{"BTC", 0, 1, 0}, Currency{"BTC", newInt(0), newInt(1), newInt(0)}},
		{"ValidPrecision", args{"BTC", 0, 0, 1}, Currency{"BTC", newInt(0), newInt(0), newInt(1)}},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCurrency(tt.args.name, tt.args.available, tt.args.crypto, tt.args.precision); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrency_IsAvailable(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name      CurrencyName
		Available ZeroInt
		Crypto    ZeroInt
		Precision ZeroInt
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default", fields{}, false},
		{"InvalidCurrency", fields{"", newInt(0), newInt(0), newInt(0)}, false},
		{"InvalidCurrencyAvailable", fields{"", newInt(0), newInt(0), newInt(0)}, false},
		{"InvalidAvailable", fields{"BTC", nil, newInt(0), newInt(0)}, false},

		{"ValidAvailable", fields{"BTC", newInt(1), newInt(0), newInt(0)}, true},
		{"ValidNotAvailable", fields{"BTC", newInt(0), newInt(0), newInt(0)}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &Currency{
				Name:      tt.fields.Name,
				Available: tt.fields.Available,
			}
			if got := p.IsAvailable(); got != tt.want {
				t.Errorf("Currency.IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrency_IsCrypto(t *testing.T) {
	type fields struct {
		Name      CurrencyName
		Available ZeroInt
		Crypto    ZeroInt
		Precision ZeroInt
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default", fields{}, false},
		{"InvalidCurrency", fields{"", newInt(0), newInt(0), newInt(0)}, false},
		{"InvalidCurrencyCrypto", fields{"", newInt(0), newInt(0), newInt(0)}, false},
		{"InvalidCrypto", fields{"BTC", newInt(0), nil, newInt(0)}, false},

		{"ValidCrypto", fields{"BTC", newInt(0), newInt(1), newInt(0)}, true},
		{"ValidNotCrypto", fields{"BTC", newInt(0), newInt(0), newInt(0)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Currency{
				Name:      tt.fields.Name,
				Available: tt.fields.Available,
				Crypto:    tt.fields.Crypto,
				Precision: tt.fields.Precision,
			}
			if got := p.IsCrypto(); got != tt.want {
				t.Errorf("Currency.IsCrypto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrency_DisplayPrecision(t *testing.T) {
	type fields struct {
		Name      CurrencyName
		Available ZeroInt
		Crypto    ZeroInt
		Precision ZeroInt
	}
	tests := []struct {
		name   string
		fields fields
		want   Int
	}{
		{"Default", fields{}, Int(0)},
		{"InvalidCurrency", fields{"", newInt(0), newInt(0), newInt(0)}, Int(0)},
		{"InvalidCurrencyDisplayPrecision", fields{"", newInt(0), newInt(0), newInt(0)}, Int(0)},
		{"InvalidDisplayPrecision", fields{"BTC", newInt(0), newInt(0), nil}, Int(0)},

		{"ValidDisplayPrecision", fields{"BTC", newInt(0), newInt(0), newInt(1)}, Int(1)},
		{"ValidNotDisplayPrecision", fields{"BTC", newInt(0), newInt(0), newInt(0)}, Int(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Currency{
				Name:      tt.fields.Name,
				Available: tt.fields.Available,
				Crypto:    tt.fields.Crypto,
				Precision: tt.fields.Precision,
			}
			if got := p.DisplayPrecision(); got != tt.want {
				t.Errorf("Currency.DisplayPrecision() = %v, want %v", got, tt.want)
			}
		})
	}
}
