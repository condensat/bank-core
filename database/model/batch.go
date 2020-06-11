// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type BatchID ID
type BatchData String

type Batch struct {
	ID        BatchID   `gorm:"primary_key"`
	Timestamp time.Time `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	Data      BatchData `gorm:"type:blob;not null;default:'{}'"` // Batch data
}
