// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type SwapInfoID ID
type SwapStatus String
type Payload String

const (
	SwapStatusProposed  SwapStatus = "proposed"
	SwapStatusAccepted  SwapStatus = "accepted"
	SwapStatusFinalized SwapStatus = "finalized"
	SwapStatusCompleted SwapStatus = "completed"
	SwapStatusCanceled  SwapStatus = "canceled"
)

type SwapInfo struct {
	ID        SwapInfoID `gorm:"primary_key"`
	Timestamp time.Time  `gorm:"index;not null;type:timestamp"` // Creation timestamp
	SwapID    SwapID     `gorm:"index;not null"`                // [FK] Reference to Swap table
	Status    SwapStatus `gorm:"index;not null;size:16"`        // SwapStatus [proposed, accepted, finalized, completed, canceled]
	Payload   Payload    `gorm:"type:blob;not null"`            // Payload swap data
}
