// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

// AssetIcon from https://assets.blockstream.info/icons.json
type AssetIcon struct {
	AssetID    AssetID   `gorm:"unique_index;not null"`         // [FK] Reference to Asset table
	LastUpdate time.Time `gorm:"index;not null;type:timestamp"` // Last update timestamp
	Data       []byte    `gorm:"type:MEDIUMBLOB;default:null"`  // Decoded data byte
}
