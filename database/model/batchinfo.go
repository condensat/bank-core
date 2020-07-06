// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"errors"
	"time"
)

type BatchInfoID ID
type BatchStatus String
type BatchInfoData Data

const (
	BatchStatusCreated    BatchStatus = "created"
	BatchStatusReady      BatchStatus = "ready"
	BatchStatusProcessing BatchStatus = "processing"
	BatchStatusConfirmed  BatchStatus = "confirmed"
	BatchStatusSettled    BatchStatus = "settled"
	BatchStatusCanceled   BatchStatus = "canceled"

	BatchInfoCrypto DataType = "crypto"
)

var (
	ErrInvalidDataType = errors.New("Invalid DataType")
)

type BatchInfo struct {
	ID        BatchInfoID   `gorm:"primary_key"`
	Timestamp time.Time     `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	BatchID   BatchID       `gorm:"index;not null"`                  // [FK] Reference to Batch table
	Status    BatchStatus   `gorm:"index;not null;size:16"`          // BatchStatus [created, processing, completed, canceled]
	Type      DataType      `gorm:"index;not null;size:16"`          // DataType [crypto]
	Data      BatchInfoData `gorm:"type:blob;not null;default:'{}'"` // BatchInfo data
}

// BatchInfoCryptoData data type for BatchInfo crypto
type BatchInfoCryptoData struct {
	TxID   String `json:"txid,omitempty"`
	Height Int    `json:"height,omitempty"`
}

func (p *BatchInfo) CryptoData() (BatchInfoCryptoData, error) {
	switch p.Type {

	case BatchInfoCrypto:
		var data BatchInfoCryptoData
		err := DecodeData(&data, Data(p.Data))
		return data, err

	default:
		return BatchInfoCryptoData{}, ErrInvalidDataType
	}
}
