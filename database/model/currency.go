// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type Currency struct {
	Name      string `gorm:"primary_key;type:varchar(16)"` // [PK] Currency
	Available *int   `gorm:"default:0;not null"`
}

func NewCurrency(name string, available int) Currency {
	return Currency{
		Name:      name,
		Available: &available,
	}
}

func (p *Currency) IsAvailable() bool {
	return p.Available != nil && *p.Available > 0
}
