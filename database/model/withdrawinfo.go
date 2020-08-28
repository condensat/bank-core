// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type WithdrawInfoID ID
type WithdrawStatus String
type WithdrawInfoData String

const (
	WithdrawStatusCreated    WithdrawStatus = "created"
	WithdrawStatusProcessing WithdrawStatus = "processing"
	WithdrawStatusSettled    WithdrawStatus = "settled"
	WithdrawStatusCanceling  WithdrawStatus = "canceling"
	WithdrawStatusCanceled   WithdrawStatus = "canceled"
)

type WithdrawInfo struct {
	ID         WithdrawInfoID   `gorm:"primary_key"`
	Timestamp  time.Time        `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	WithdrawID WithdrawID       `gorm:"index;not null"`                  // [FK] Reference to Withdraw table
	Status     WithdrawStatus   `gorm:"index;not null;size:16"`          // WithdrawStatus [created, processing, completed, canceled]
	Data       WithdrawInfoData `gorm:"type:blob;not null;default:'{}'"` // WithdrawInfo data
}
