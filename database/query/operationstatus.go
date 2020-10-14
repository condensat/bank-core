// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"
	"time"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidOperationStatus = errors.New("Invalid OperationStatus")
)

// AddOrUpdateOperationStatus
func AddOrUpdateOperationStatus(db database.Context, operation model.OperationStatus) (model.OperationStatus, error) {
	var result model.OperationStatus
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, database.ErrInvalidDatabase
	}

	if operation.OperationInfoID == 0 {
		return result, ErrInvalidOperationInfoID
	}

	if len(operation.State) == 0 {
		return result, ErrInvalidOperationStatus
	}

	operation.LastUpdate = time.Now().UTC().Truncate(time.Second)

	err := gdb.
		Where(model.OperationStatus{
			OperationInfoID: operation.OperationInfoID,
		}).
		Assign(operation).
		FirstOrCreate(&result).Error

	return result, err
}

type DepositInfos struct {
	Count  int
	Active int
}

func DepositsInfos(db database.Context) (DepositInfos, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return DepositInfos{}, database.ErrInvalidDatabase
	}

	var totalOperations int64
	err := gdb.Model(&model.OperationStatus{}).
		Count(&totalOperations).Error
	if err != nil {
		return DepositInfos{}, err
	}

	var activeOperations int64
	err = gdb.Model(&model.OperationStatus{}).
		Where("state <> ?", "settled").
		Count(&activeOperations).Error
	if err != nil {
		return DepositInfos{}, err
	}

	return DepositInfos{
		Count:  int(totalOperations),
		Active: int(activeOperations),
	}, nil
}

// GetOperationStatus
func GetOperationStatus(db database.Context, operationInfoID model.OperationInfoID) (model.OperationStatus, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.OperationStatus{}, database.ErrInvalidDatabase
	}

	if operationInfoID == 0 {
		return model.OperationStatus{}, ErrInvalidOperationInfoID
	}

	var result model.OperationStatus
	err := gdb.
		Where(model.OperationStatus{
			OperationInfoID: operationInfoID,
		}).
		First(&result).Error
	if err != nil {
		return model.OperationStatus{}, err
	}

	return result, nil
}

func FindActiveOperationStatus(db database.Context) ([]model.OperationStatus, error) {
	gdb := db.DB().(*gorm.DB)

	var list []*model.OperationStatus
	err := gdb.
		Where("state <> ?", "settled").
		Or("accounted <> ?", "settled").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertOperationStatusList(list), nil
}

func FindActiveOperationInfo(db database.Context) ([]model.OperationInfo, error) {
	gdb := db.DB().(*gorm.DB)

	subQueryState := gdb.Model(&model.OperationStatus{}).
		Select("operation_info_id").
		Where("state <> ?", "settled").
		Or("accounted <> ?", "settled").
		SubQuery()

	var list []*model.OperationInfo
	err := gdb.Model(&model.OperationInfo{}).
		Joins("JOIN (?) AS os ON operation_info.id = os.operation_info_id", subQueryState).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertOperationInfoList(list), nil
}

func convertOperationStatusList(list []*model.OperationStatus) []model.OperationStatus {
	var result []model.OperationStatus
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *curr)
	}

	return result[:]
}
