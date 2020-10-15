// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"

	"github.com/condensat/bank-core/messaging"
)

type ProcessInfo struct {
	ID        uint64    `gorm:"primary_key"`
	Timestamp time.Time `gorm:"index;not null"`
	AppName   string    `gorm:"index;not null"`
	Hostname  string    `gorm:"index;not null"`
	PID       int       `gorm:"not null"`

	MemAlloc      uint64 `gorm:"not null"`
	MemTotalAlloc uint64 `gorm:"not null"`
	MemSys        uint64 `gorm:"not null"`
	MemLookups    uint64 `gorm:"not null"`

	NumCPU       uint64  `gorm:"not null"`
	NumGoroutine uint64  `gorm:"not null"`
	NumCgoCall   uint64  `gorm:"not null"`
	CPUUsage     float64 `gorm:"not null"`
}

func (p *ProcessInfo) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *ProcessInfo) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}
