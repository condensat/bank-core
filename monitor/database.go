// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/monitor/common"
	"github.com/jinzhu/gorm"
)

func AddProcessInfo(ctx context.Context, processInfo *common.ProcessInfo) error {
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return errors.New("Wrong database")
	}

	return db.Create(&processInfo).Error
}

func ListServices(ctx context.Context, since time.Duration) ([]string, error) {
	var result []string
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return result, errors.New("Wrong database")
	}

	now := time.Now().UTC()
	distinctAppName := fmt.Sprintf("distinct (%s)", gorm.ToColumnName("AppName"))

	var list []*common.ProcessInfo
	err := db.Select(distinctAppName).
		Where("timestamp BETWEEN ? AND ?", now.Add(-since), now).
		Find(&list).Error
	if err != nil {
		return result, err
	}

	for _, entry := range list {
		result = append(result, entry.AppName)
	}

	return result, nil
}

func LastServicesStatus(ctx context.Context) ([]common.ProcessInfo, error) {
	var result []common.ProcessInfo
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return nil, errors.New("Wrong database")
	}

	subQuery := db.Model(&common.ProcessInfo{}).
		Select("MAX(id) as id, MAX(timestamp) AS last").
		Where("timestamp >= DATE_SUB(NOW(), INTERVAL 3 MINUTE)").
		Group("app_name, hostname").
		SubQuery()

	var list []*common.ProcessInfo
	err := db.Joins("RIGHT JOIN (?) AS t1 ON process_info.id = t1.id AND timestamp = t1.last", subQuery).
		Order("app_name ASC, hostname ASC, timestamp DESC").
		Find(&list).Error

	if err != nil {
		return nil, err
	}

	for _, entry := range list {
		result = append(result, *entry)
	}

	return result, nil
}
