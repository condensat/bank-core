// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"
	"time"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/encoding"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidBatchInfoID       = errors.New("Invalid BatchInfoID")
	ErrInvalidBatchStatus       = errors.New("Invalid BatchInfo Status")
	ErrInvalidBatchInfoDataType = errors.New("Invalid BatchInfo DataType")
)

func AddBatchInfo(db database.Context, batchID model.BatchID, status model.BatchStatus, dataType encoding.DataType, data model.BatchInfoData) (model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.BatchInfo{}, database.ErrInvalidDatabase
	}

	if batchID == 0 {
		return model.BatchInfo{}, ErrInvalidBatchID
	}
	if len(status) == 0 {
		return model.BatchInfo{}, ErrInvalidBatchStatus
	}
	if len(dataType) == 0 {
		return model.BatchInfo{}, ErrInvalidBatchInfoDataType
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.BatchInfo{
		Timestamp: timestamp,
		BatchID:   batchID,
		Status:    status,
		Type:      dataType,
		Data:      data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.BatchInfo{}, err
	}

	return result, nil

}

func GetBatchInfo(db database.Context, ID model.BatchInfoID) (model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.BatchInfo{}, database.ErrInvalidDatabase
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

func GetBatchHistory(db database.Context, batchID model.BatchID) ([]model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, database.ErrInvalidDatabase
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

func GetBatchInfoByStatusAndType(db database.Context, status model.BatchStatus, dataType encoding.DataType) ([]model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}

	if len(status) == 0 {
		return nil, ErrInvalidWithdrawStatus
	}
	if len(dataType) == 0 {
		return nil, ErrInvalidBatchInfoDataType
	}

	var list []*model.BatchInfo
	err := gdb.
		Where(model.BatchInfo{
			Status: status,
			Type:   dataType,
		}).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertBatchInfoList(list), nil
}

func GetLastBatchInfo(db database.Context, batchID model.BatchID) (model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.BatchInfo{}, database.ErrInvalidDatabase
	}

	if batchID == 0 {
		return model.BatchInfo{}, ErrInvalidBatchID
	}

	subQueryLast := gdb.Model(&model.BatchInfo{}).
		Select("MAX(id)").
		Group("batch_id").
		SubQuery()

	var result model.BatchInfo
	err := gdb.Model(&model.BatchInfo{}).
		Where("batch_info.id IN (?)", subQueryLast).
		Where(model.BatchInfo{
			BatchID: batchID,
		}).First(&result).Error

	if err != nil {
		return model.BatchInfo{}, err
	}

	return result, nil
}

func GetLastBatchInfoByStatusAndNetwork(db database.Context, status model.BatchStatus, network model.BatchNetwork) ([]model.BatchInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}

	if len(status) == 0 {
		return nil, ErrInvalidWithdrawStatus
	}
	if len(network) == 0 {
		return nil, ErrInvalidNetwork
	}

	subQueryLast := gdb.Model(&model.BatchInfo{}).
		Select("MAX(id)").
		Group("batch_id").
		SubQuery()

	subQueryNetwork := gdb.Model(&model.Batch{}).
		Select("id").
		Where(model.Batch{
			Network: network,
		}).
		SubQuery()

	var list []*model.BatchInfo
	err := gdb.Model(&model.BatchInfo{}).
		Joins("JOIN (?) AS b ON batch_info.batch_id = b.id", subQueryNetwork).
		Where("batch_info.id IN (?)", subQueryLast).
		Where(model.BatchInfo{
			Status: status,
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

func BatchPagingCount(db database.Context, countByPage int) (int, error) {
	if countByPage <= 0 {
		countByPage = 1
	}

	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result int
		err := gdb.
			Model(&model.BatchInfo{}).
			Group("batch_id").
			Count(&result).Error
		var partialPage int
		if result%countByPage > 0 {
			partialPage = 1
		}
		return result/countByPage + partialPage, err

	default:
		return 0, database.ErrInvalidDatabase
	}
}

func BatchPage(db database.Context, batchID model.BatchID, countByPage int) ([]model.BatchID, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if countByPage <= 0 {
			countByPage = 1
		}

		subQueryLast := gdb.Model(&model.BatchInfo{}).
			Select("MAX(id)").
			Group("batch_id").
			SubQuery()

		var list []*model.BatchInfo
		err := gdb.Model(&model.BatchInfo{}).
			Where("batch_info.id IN (?)", subQueryLast).
			Where("id >= ?", batchID).
			Order("batch_id ASC").
			Limit(countByPage).
			Find(&list).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		return convertBatchInfoIDs(list), nil

	default:
		return nil, database.ErrInvalidDatabase
	}
}

func convertBatchInfoIDs(list []*model.BatchInfo) []model.BatchID {
	var result []model.BatchID
	for _, curr := range list {
		if curr != nil {
			result = append(result, curr.BatchID)
		}
	}

	return result[:]
}
