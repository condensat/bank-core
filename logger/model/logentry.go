// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
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
	Method    string `gorm:"index;type:varchar(48)"`
	Error     string `gorm:"index;type:varchar(256)"`

	Message string `gorm:"type:varchar(256)"`
	Data    string `gorm:"type:json"`
}

func TxAddLogEntries(db *gorm.DB, entries []*LogEntry) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			err := tx.Create(entry).Error
			if err != nil {
				// do not rollback tx
				fmt.Printf("TxAddLogEntries failed. %s", err)
				continue
			}
		}

		return nil
	})
}

type LogInfo struct {
	Warnings int
	Errors   int
	Panics   int
}

func countLogByLevel(db *gorm.DB, level string) (int64, error) {
	var result int64
	err := db.Model(&LogEntry{}).
		Where(LogEntry{
			Level: level,
		}).
		Where("timestamp >= now() - interval 1 day").
		Count(&result).Error

	return result, err
}

func LogsInfo(db *gorm.DB) (LogInfo, error) {
	var result LogInfo

	err := db.Transaction(func(tx *gorm.DB) error {
		// Warning count
		count, err := countLogByLevel(tx, "warning")
		if err != nil {
			return err
		}
		result.Warnings = int(count)

		// Error count
		count, err = countLogByLevel(tx, "error")
		if err != nil {
			return err
		}
		result.Errors = int(count)

		// Panic count
		count, err = countLogByLevel(tx, "panic")
		if err != nil {
			return err
		}
		result.Panics = int(count)

		return nil
	})

	return result, err
}
