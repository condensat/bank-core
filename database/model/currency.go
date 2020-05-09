// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type CurrencyName String
type CurrencyAvailable ZeroInt

type Currency struct {
	Name        CurrencyName `gorm:"primary_key;type:varchar(16)"` // [PK] Currency
	DisplayName CurrencyName `gorm:"type:varchar(32)"`
	Type        ZeroInt      `gorm:"default:0;not null"` // currencyType [Fiat=0, CryptoNative=1, CryptoAsset=2]
	Available   ZeroInt      `gorm:"default:0;not null"`
	Crypto      ZeroInt      `gorm:"default:0;not null"`
	Precision   ZeroInt      `gorm:"default:0;not null"`
	AutoCreate  bool         `gorm:"default:false"` // Automatic creation for accounts
}

func NewCurrency(name, displayName CurrencyName, currencyType, available, crypto, precision Int) Currency {
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
		Name:        name,
		DisplayName: displayName,
		Type:        ZeroInt(&currencyType),
		Available:   ZeroInt(&available),
		Crypto:      ZeroInt(&crypto),
		Precision:   ZeroInt(&precision),
	}
}

func (p *Currency) IsAvailable() bool {
	return len(p.Name) > 0 && p.Available != nil && *p.Available > 0
}

func (p *Currency) IsCrypto() bool {
	return len(p.Name) > 0 && p.Crypto != nil && *p.Crypto > 0
}

func (p *Currency) GetType() Int {
	if p.Type == nil {
		return 0
	}
	return *p.Type
}

func (p *Currency) DisplayPrecision() Int {
	var result Int
	if len(p.Name) > 0 && p.Precision != nil {
		result = *p.Precision
	}

	return result
}
