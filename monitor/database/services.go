// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"fmt"
	"time"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/monitor/database/model"

	"github.com/jinzhu/gorm"
)

func ListServices(db database.Context, since time.Duration) ([]string, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		panic("Invalid db")
	}

	now := time.Now().UTC()
	distinctAppName := fmt.Sprintf("distinct (%s)", gorm.ToColumnName("AppName"))

	var list []*model.ProcessInfo
	err := gdb.Select(distinctAppName).
		Where("timestamp BETWEEN ? AND ?", now.Add(-since), now).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var result []string
	for _, entry := range list {
		result = append(result, entry.AppName)
	}

	return result, nil
}

func LastServicesStatus(db database.Context) ([]model.ProcessInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		panic("Invalid db")
	}

	subQuery := gdb.Model(&model.ProcessInfo{}).
		Select("MAX(id) as id, MAX(timestamp) AS last").
		Where("timestamp >= DATE_SUB(NOW(), INTERVAL 3 MINUTE)").
		Group("app_name, hostname").
		SubQuery()

	var list []*model.ProcessInfo
	err := gdb.Joins("RIGHT JOIN (?) AS t1 ON process_info.id = t1.id AND timestamp = t1.last", subQuery).
		Order("app_name ASC, hostname ASC, timestamp DESC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var result []model.ProcessInfo
	for _, entry := range list {
		result = append(result, *entry)
	}

	return result, nil
}

func LastServiceHistory(db database.Context, appName string, from, to time.Time, step time.Duration, round time.Duration) ([]model.ProcessInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		panic("Invalid db")
	}

	tsFrom := from.UnixNano() / int64(time.Second)
	tsTo := to.UnixNano() / int64(time.Second)

	subQuery := gdb.Model(&model.ProcessInfo{}).
		Select("MAX(id) AS id, FLOOR(UNIX_TIMESTAMP(timestamp)/(?)) AS timekey", step/time.Second).
		Where("app_name=?", appName).
		Where("timestamp BETWEEN FROM_UNIXTIME(?) AND FROM_UNIXTIME(?)", tsFrom, tsTo).
		Group("timekey, hostname").
		SubQuery()

	var list []*model.ProcessInfo
	err := gdb.Joins("RIGHT JOIN (?) AS t1 ON process_info.id = t1.id", subQuery).
		Order("timestamp, hostname DESC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var result []model.ProcessInfo
	for _, entry := range list {
		entry.Timestamp = entry.Timestamp.Round(round)
		result = append(result, *entry)
	}

	return result, nil
}
