// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "time"

type BlockID ID

const MemPoolBlockID = BlockID(1)

type CryptoAddress struct {
	ID            ID         `gorm:"primary_key"`                    // [PK] CryptoAddress
	AccountID     AccountID  `gorm:"index;not null"`                 // [FK] Reference to Account table
	PublicAddress String     `gorm:"unique_index;not null;size:128"` // CryptoAddress public key, non mutable
	Chain         String     `gorm:"index;not null;size:16"`         // CryptoAddress chain, non mutable
	CreationDate  *time.Time `gorm:"index;not null"`                 // CryptoAddress creation date, non mutable
	FirstBlockId  BlockID    `gorm:"index;not null"`                 // Block height of the first transaction
}

func (p *CryptoAddress) IsUsed() bool {
	return p.FirstBlockId > 0
}

func (p *CryptoAddress) Confirmations(height BlockID) int {
	if !p.IsUsed() {
		return 0
	}
	if p.FirstBlockId <= MemPoolBlockID {
		return 0
	}
	if p.FirstBlockId > height {
		return 0
	}
	return 1 + int(height) - int(p.FirstBlockId)
}
