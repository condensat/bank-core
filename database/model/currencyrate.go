// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type CurrencyRateID ID
type CurrencyRateSource String
type CurrencyRateValue Float

type CurrencyRate struct {
	ID        CurrencyRateID     `gorm:"primary_key"`
	Timestamp time.Time          `gorm:"index;not null;type:timestamp"`
	Source    CurrencyRateSource `gorm:"index;not null;type:varchar(16)"`
	Base      CurrencyName       `gorm:"index;not null;type:varchar(16)"`
	Name      CurrencyName       `gorm:"index;not null;type:varchar(16)"`
	Rate      CurrencyRateValue  `gorm:"not null"`
}
