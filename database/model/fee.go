// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type FeeID ID
type FeeData String

type Fee struct {
	ID         FeeID      `gorm:"primary_key"`
	WithdrawID WithdrawID `gorm:"unique_index;not null"`           // [FK] Reference to Withdraw table
	Amount     ZeroFloat  `gorm:"default:0;not null"`              // Operation amount (can be negative)
	Data       FeeData    `gorm:"type:blob;not null;default:'{}'"` // Fee data
}
