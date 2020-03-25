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
	}
	tests := []struct {
		name string
		args args
		want Currency
	}{
		{"Default", args{"", 0}, Currency{}},
		{"InvalidCurrency", args{"", 1}, Currency{}},
		{"InvalidAvailable", args{"BTC", -1}, Currency{}},

		{"Valid", args{"BTC", 0}, Currency{"BTC", newInt(0)}},
		{"ValidAvailable", args{"BTC", 1}, Currency{"BTC", newInt(1)}},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCurrency(tt.args.name, tt.args.available); !reflect.DeepEqual(got, tt.want) {
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
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default", fields{}, false},
		{"InvalidCurrency", fields{"", newInt(0)}, false},
		{"InvalidCurrencyAvailable", fields{"", newInt(1)}, false},
		{"InvalidAvailable", fields{"BTC", nil}, false},

		{"ValidAvailable", fields{"BTC", newInt(1)}, true},
		{"ValidNotAvailable", fields{"BTC", newInt(0)}, false},
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
