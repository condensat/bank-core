// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type BatchID ID
type BatchData String
type BatchNetwork String

const (
	BatchNetworkSepa  BatchNetwork = "sepa"
	BatchNetworkSwift BatchNetwork = "swift"
	BatchNetworkCard  BatchNetwork = "card"

	BatchNetworkBitcoin          BatchNetwork = "bitcoin"
	BatchNetworkBitcoinTestnet   BatchNetwork = "bitcoin-testnet"
	BatchNetworkBitcoinLiquid    BatchNetwork = "liquid"
	BatchNetworkBitcoinLightning BatchNetwork = "lightning"
)

type Batch struct {
	ID        BatchID      `gorm:"primary_key"`
	Timestamp time.Time    `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	Network   BatchNetwork `gorm:"index;not null;size:24"`          // Network [sepa, swift, card, bitcoin, bitcoin-testnet, liquid, lightning]
	Data      BatchData    `gorm:"type:blob;not null;default:'{}'"` // Batch data
}
