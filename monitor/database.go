// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

import (
	"context"
	"errors"

	"github.com/condensat/bank-core/appcontext"
	"github.com/jinzhu/gorm"
)

func AddProcessInfo(ctx context.Context, processInfo *ProcessInfo) error {
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return errors.New("Wrong database")
	}

	return db.Create(&processInfo).Error
}
