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
	ErrInvalidWithdrawInfoID   = errors.New("Invalid WithdrawInfoID")
	ErrInvalidWithdrawStatus   = errors.New("Invalid WithdrawInfo Status")
	ErrInvalidWithdrawInfoData = errors.New("Invalid WithdrawInfo Data")
)

func AddWithdrawInfo(db bank.Database, withdrawID model.WithdrawID, status model.WithdrawStatus, data model.WithdrawInfoData) (model.WithdrawInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.WithdrawInfo{}, errors.New("Invalid appcontext.Database")
	}

	if withdrawID == 0 {
		return model.WithdrawInfo{}, ErrInvalidWithdrawID
	}
	if len(status) == 0 {
		return model.WithdrawInfo{}, ErrInvalidWithdrawStatus
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.WithdrawInfo{
		Timestamp:  timestamp,
		WithdrawID: withdrawID,
		Status:     status,
		Data:       data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.WithdrawInfo{}, err
	}

	return result, nil

}

func GetWithdrawInfo(db bank.Database, ID model.WithdrawInfoID) (model.WithdrawInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.WithdrawInfo{}, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return model.WithdrawInfo{}, ErrInvalidWithdrawInfoID
	}

	var result model.WithdrawInfo
	err := gdb.
		Where(&model.WithdrawInfo{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.WithdrawInfo{}, err
	}

	return result, nil
}

func GetLastWithdrawInfo(db bank.Database, withdrawID model.WithdrawID) (model.WithdrawInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.WithdrawInfo{}, errors.New("Invalid appcontext.Database")
	}

	if withdrawID == 0 {
		return model.WithdrawInfo{}, ErrInvalidWithdrawInfoID
	}

	var result model.WithdrawInfo
	err := gdb.
		Where(&model.WithdrawInfo{WithdrawID: withdrawID}).
		Last(&result).Error
	if err != nil {
		return model.WithdrawInfo{}, err
	}

	return result, nil

}

func GetWithdrawHistory(db bank.Database, withdrawID model.WithdrawID) ([]model.WithdrawInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if withdrawID == 0 {
		return nil, ErrInvalidWithdrawID
	}

	var list []*model.WithdrawInfo
	err := gdb.
		Where(model.WithdrawInfo{
			WithdrawID: withdrawID,
		}).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertWithdrawInfoList(list), nil
}

func ListCancelingWithdrawsAccountOperations(db bank.Database) ([]model.AccountOperation, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return nil, ErrInvalidDatabase
	}

	subQueryLastWitdrawInfo := gdb.Model(&model.WithdrawInfo{}).
		Select("MAX(id)").
		Group("withdraw_id").
		SubQuery()

	var list []*model.AccountOperation
	err := gdb.
		Joins("JOIN (withdraw_info) ON account_operation.reference_id = withdraw_info.withdraw_id").
		Where("withdraw_info.id IN (?)", subQueryLastWitdrawInfo).
		Where(model.AccountOperation{OperationType: model.OperationTypeTransfer}).
		Where(model.WithdrawInfo{Status: model.WithdrawStatusCanceling}).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertAccountOperationList(list), nil
}

func convertWithdrawInfoList(list []*model.WithdrawInfo) []model.WithdrawInfo {
	var result []model.WithdrawInfo
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}

func WithdrawPagingCount(db bank.Database, countByPage int) (int, error) {
	if countByPage <= 0 {
		countByPage = 1
	}

	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result int
		err := gdb.
			Model(&model.WithdrawInfo{}).
			Group("withdraw_id").
			Count(&result).Error
		var partialPage int
		if result%countByPage > 0 {
			partialPage = 1
		}
		return result/countByPage + partialPage, err

	default:
		return 0, ErrInvalidDatabase
	}
}

func WithdrawPage(db bank.Database, withdrawID model.WithdrawID, countByPage int) ([]model.WithdrawID, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if countByPage <= 0 {
			countByPage = 1
		}

		subQueryLast := gdb.Model(&model.WithdrawInfo{}).
			Select("MAX(id)").
			Group("withdraw_id").
			SubQuery()

		var list []*model.WithdrawInfo
		err := gdb.Model(&model.WithdrawInfo{}).
			Where("withdraw_info.id IN (?)", subQueryLast).
			Where("id >= ?", withdrawID).
			Order("withdraw_id ASC").
			Limit(countByPage).
			Find(&list).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		return convertWithdrawInfoIDs(list), nil

	default:
		return nil, ErrInvalidDatabase
	}
}

func convertWithdrawInfoIDs(list []*model.WithdrawInfo) []model.WithdrawID {
	var result []model.WithdrawID
	for _, curr := range list {
		if curr != nil {
			result = append(result, curr.WithdrawID)
		}
	}

	return result[:]
}
