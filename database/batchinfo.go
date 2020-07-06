// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidBatchInfoID   = errors.New("Invalid BatchInfoID")
	ErrInvalidBatchStatus   = errors.New("Invalid BatchInfo Status")
	ErrInvalidBatchInfoData = errors.New("Invalid BatchInfo Data")
)

func AddBatchInfo(db bank.Database, batchID model.BatchID, status model.BatchStatus, data model.BatchInfoData) (model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.BatchInfo{}, errors.New("Invalid appcontext.Database")
	}

	if batchID == 0 {
		return model.BatchInfo{}, ErrInvalidBatchID
	}
	if len(status) == 0 {
		return model.BatchInfo{}, ErrInvalidBatchStatus
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.BatchInfo{
		Timestamp: timestamp,
		BatchID:   batchID,
		Status:    status,
		Data:      data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.BatchInfo{}, err
	}

	return result, nil

}

func GetBatchInfo(db bank.Database, ID model.BatchInfoID) (model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.BatchInfo{}, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return model.BatchInfo{}, ErrInvalidBatchInfoID
	}

	var result model.BatchInfo
	err := gdb.
		Where(&model.BatchInfo{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.BatchInfo{}, err
	}

	return result, nil
}

func GetBatchHistory(db bank.Database, batchID model.BatchID) ([]model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if batchID == 0 {
		return nil, ErrInvalidBatchID
	}

	var list []*model.BatchInfo
	err := gdb.
		Where(model.BatchInfo{
			BatchID: batchID,
		}).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertBatchInfoList(list), nil
}

func convertBatchInfoList(list []*model.BatchInfo) []model.BatchInfo {
	var result []model.BatchInfo
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}
