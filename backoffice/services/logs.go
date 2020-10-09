// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger/model"

	"github.com/jinzhu/gorm"
)

type LogStatus struct {
	Warnings int `json:"warning"`
	Errors   int `json:"errors"`
	Panics   int `json:"panics"`
}

func FetchLogStatus(ctx context.Context) (LogStatus, error) {
	db := appcontext.Database(ctx)

	logsInfo, err := model.LogsInfo(db.DB().(*gorm.DB))

	return LogStatus{
		Warnings: logsInfo.Warnings,
		Errors:   logsInfo.Errors,
		Panics:   logsInfo.Panics,
	}, err
}
