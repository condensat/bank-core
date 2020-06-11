// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type SwapID ID
type SwapType String

const (
	SwapTypeBid SwapType = "bid"
	SwapTypeAsk SwapType = "ask"

	SwapTypeInternalBid SwapType = "internal_bid"
	SwapTypeInternalAsk SwapType = "internal_ask"
)

type Swap struct {
	ID              SwapID          `gorm:"primary_key"`
	Timestamp       time.Time       `gorm:"index;not null;type:timestamp"` // Creation timestamp
	ValidUntil      time.Time       `gorm:"index;not null;type:timestamp"` // Valid Until
	Type            SwapType        `gorm:"index;not null;size:16"`        // SwapType [bid, ask]
	CryptoAddressID CryptoAddressID `gorm:"index;not null"`                // [FK] Reference to CryptoAddress table
	DebitAsset      AssetID         `gorm:"index;not null"`                // [FK] Reference to Asset table for Debit
	DebitAmount     Float           `gorm:"default:0;not null"`            // DebitAmount (strictly positive)
	CreditAsset     AssetID         `gorm:"index;not null"`                // [FK] Reference to Asset table for Credit
	CreditAmount    Float           `gorm:"default:0;not null"`            // CreditAmount (strictly positive)
}
