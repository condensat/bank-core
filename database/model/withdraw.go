// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type WithdrawID ID
type BatchMode String
type WithdrawData String

const (
	BatchModeInstant BatchMode = "instant"
	BatchModeFast    BatchMode = "fast"
	BatchModeNormal  BatchMode = "normal"
	BatchModeSlow    BatchMode = "slow"
)

type Withdraw struct {
	ID        WithdrawID   `gorm:"primary_key"`
	Timestamp time.Time    `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	From      AccountID    `gorm:"index;not null"`                  // [FK] Reference to Account table
	To        AccountID    `gorm:"index;not null"`                  // [FK] Reference to Account table
	Amount    ZeroFloat    `gorm:"default:0;not null"`              // Operation amount (can not be negative)
	Batch     BatchMode    `gorm:"index;not null;size:16"`          // BatchMode [instant, fast, normal, slow]
	Data      WithdrawData `gorm:"type:blob;not null;default:'{}'"` // Withdraw data
}
