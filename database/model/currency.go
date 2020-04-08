// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type CurrencyName String
type CurrencyAvailable ZeroInt

type Currency struct {
	Name      CurrencyName `gorm:"primary_key;type:varchar(16)"` // [PK] Currency
	Available ZeroInt      `gorm:"default:0;not null"`
	Crypto    ZeroInt      `gorm:"default:0;not null"`
	Precision ZeroInt      `gorm:"default:2;not null"`
}

func NewCurrency(name CurrencyName, available, crypto, precision Int) Currency {
	if len(name) == 0 {
		return Currency{}
	}
	if available < 0 {
		available = 0
	}
	if crypto < 0 {
		crypto = 0
	}
	if precision < 0 {
		precision = 2
	}

	return Currency{
		Name:      name,
		Available: ZeroInt(&available),
		Crypto:    ZeroInt(&crypto),
		Precision: ZeroInt(&precision),
	}
}

func (p *Currency) IsAvailable() bool {
	return len(p.Name) > 0 && p.Available != nil && *p.Available > 0
}

func (p *Currency) IsCrypto() bool {
	return len(p.Name) > 0 && p.Crypto != nil && *p.Crypto > 0
}

func (p *Currency) DisplayPrecision() Int {
	var result Int
	if len(p.Name) > 0 && p.Precision != nil {
		result = *p.Precision
	}

	return result
}
