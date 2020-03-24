// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type Account struct {
	ID           uint64 `gorm:"primary_key"`                     // [PK] Account
	UserID       uint64 `gorm:"index;not null"`                  // [FK] Reference to User table
	CurrencyName string `gorm:"index;not null;type:varchar(16)"` // [FK] Reference to Currency table
	Name         string `gorm:"index;not null"`                  // [U] Unique Account name for User and Currency
}
