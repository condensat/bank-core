// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"
)

type BatchInfoID ID
type BatchStatus String
type BatchInfoData String

const (
	BatchStatusCreated    BatchStatus = "created"
	BatchStatusProcessing BatchStatus = "processing"
	BatchStatusSettled    BatchStatus = "settled"
	BatchStatusCanceled   BatchStatus = "canceled"
)

type BatchInfo struct {
	ID        BatchInfoID   `gorm:"primary_key"`
	Timestamp time.Time     `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	BatchID   BatchID       `gorm:"index;not null"`                  // [FK] Reference to Batch table
	Status    BatchStatus   `gorm:"index;not null;size:16"`          // BatchStatus [created, processing, completed, canceled]
	Data      BatchInfoData `gorm:"type:blob;not null;default:'{}'"` // BatchInfo data
}
