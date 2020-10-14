// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/monitor/database/model"

	"github.com/jinzhu/gorm"
)

func AddProcessInfo(db database.Context, processInfo *model.ProcessInfo) error {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		panic("Invalid db")
	}

	return gdb.Create(&processInfo).Error
}
