// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "time"

type BlockID ID

type CryptoAddress struct {
	AccountID     AccountID  `gorm:"unique_index;not null"`         // [FK] Reference to Account table
	PublicAddress String     `gorm:"unique_index;not null;size:64"` // CryptoAddress public key, can not be reset
	CreationDate  *time.Time `gorm:"index;not null"`                // CryptoAddress creation date, non mutable
	FirstBlockId  BlockID    `gorm:"index;not null"`                // Block height of the first transaction
}
