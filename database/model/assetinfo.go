// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

// AssetInfo from https://assets.blockstream.info/
type AssetInfo struct {
	AssetID    AssetID   `gorm:"unique_index;not null"`         // [FK] Reference to Asset table
	LastUpdate time.Time `gorm:"index;not null;type:timestamp"` // Last update timestamp
	Domain     string    `gorm:"index;not null;size:253"`       // AssetDomaine name (RFC 1053)
	Name       string    `gorm:"index;not null;size:255"`       // Asset unique asset name
	Ticker     string    `gorm:"index;not null;size:5"`         // Asset unique ticker name
	Precision  uint8     `gorm:"default:0;not null"`            // Asset precision [0, 8]
}

func (p *AssetInfo) Valid() bool {
	return p.AssetID > 0 &&
		len(p.Domain) > 0 &&
		len(p.Domain) <= 253 &&
		len(p.Name) >= 2 && len(p.Name) <= 255 &&
		len(p.Ticker) >= 3 && len(p.Ticker) <= 5 &&
		p.Precision <= 8
}
