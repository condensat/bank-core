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
	Timestamp time.Time `gorm:"type:timestamp;index:timestamp_idx"`
	App       string    `gorm:"type:varchar(16);index:app_idx"`
	Level     string    `gorm:"type:varchar(16);index:level_idx"`
	Msg       string    `gorm:"type:varchar(256)"`
	Data      string    `gorm:"type:json"`
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
