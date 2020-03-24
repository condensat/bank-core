// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type CurrencyName String
type CurrencyAvailable ZeroInt

type Currency struct {
	Name      CurrencyName `gorm:"primary_key;type:varchar(16)"` // [PK] Currency
	Available ZeroInt      `gorm:"default:0;not null"`
}

func NewCurrency(name CurrencyName, available int) Currency {
	return Currency{
		Name:      name,
		Available: ZeroInt(&available),
	}
}

func (p *Currency) IsAvailable() bool {
	return p.Available != nil && *p.Available > 0
}
