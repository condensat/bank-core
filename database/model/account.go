// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type AccountID ID
type AccountName String

type Account struct {
	ID           AccountID    `gorm:"primary_key"`                     // [PK] Account
	UserID       UserID       `gorm:"index;not null"`                  // [FK] Reference to User table
	CurrencyName CurrencyName `gorm:"index;not null;type:varchar(16)"` // [FK] Reference to Currency table
	Name         AccountName  `gorm:"index;not null"`                  // [U] Unique Account name for User and Currency
}
