// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"github.com/jinzhu/gorm"
)

func Migrate(db *gorm.DB) error {
	// Automigrate all package models
	return db.AutoMigrate(
		&LogEntry{},
	).Error
}
