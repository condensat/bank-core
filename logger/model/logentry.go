// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type LogEntry struct {
	ID        uint      `gorm:"primary_key"`
	Timestamp time.Time `gorm:"index;not null;type:timestamp"`
	App       string    `gorm:"index;not null;type:varchar(16)"`
	Level     string    `gorm:"index;not null;type:varchar(16)"`

	// Optionals
	UserID    uint64 `gorm:"index"`
	SessionID string `gorm:"index;type:char(36)"` // UUID
	Method    string `gorm:"index;type:varchar(32)"`
	Error     string `gorm:"index;type:varchar(256)"`

	Message string `gorm:"type:varchar(256)"`
	Data    string `gorm:"type:json"`
}

func TxAddLogEntries(db *gorm.DB, entries []*LogEntry) error {
	tx := db.Begin()
	for _, entry := range entries {
		err := db.Create(entry).Error
		if err != nil {
			return tx.Rollback().Error
		}
	}
	return tx.Commit().Error
}
