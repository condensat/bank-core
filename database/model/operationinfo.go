// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type TxID String

type OperationInfo struct {
	ID              ID        `gorm:"primary_key;"`                  // [PK] OperationInfo
	CryptoAddressID ID        `gorm:"index;not null"`                // [FK] Reference to CryptoAddress table
	Timestamp       time.Time `gorm:"index;not null;type:timestamp"` // Creation timestamp
	TxID            TxID      `gorm:"unique_index;not null;size:64"` // Transaction ID
	Data            String    `gorm:"type:json;not null"`            // Specific operation json data
}
