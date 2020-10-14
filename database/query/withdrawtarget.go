// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidWithdrawTargetID   = errors.New("Invalid WithdrawTargetID")
	ErrInvalidWithdrawTargetData = errors.New("Invalid WithdrawTarget Data")
)

func AddWithdrawTarget(db database.Context, withdrawID model.WithdrawID, dataType model.WithdrawTargetType, data model.WithdrawTargetData) (model.WithdrawTarget, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.WithdrawTarget{}, database.ErrInvalidDatabase
	}

	if withdrawID == 0 {
		return model.WithdrawTarget{}, ErrInvalidWithdrawID
	}
	if len(dataType) == 0 {
		return model.WithdrawTarget{}, model.ErrInvalidDataType
	}

	result := model.WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       dataType,
		Data:       data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.WithdrawTarget{}, err
	}

	return result, nil

}

func GetWithdrawTarget(db database.Context, ID model.WithdrawTargetID) (model.WithdrawTarget, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.WithdrawTarget{}, database.ErrInvalidDatabase
	}

	if ID == 0 {
		return model.WithdrawTarget{}, ErrInvalidWithdrawTargetID
	}

	var result model.WithdrawTarget
	err := gdb.
		Where(&model.WithdrawTarget{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.WithdrawTarget{}, err
	}

	return result, nil
}

func GetWithdrawTargetByWithdrawID(db database.Context, withdrawID model.WithdrawID) (model.WithdrawTarget, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.WithdrawTarget{}, database.ErrInvalidDatabase
	}

	if withdrawID == 0 {
		return model.WithdrawTarget{}, ErrInvalidWithdrawID
	}

	var result model.WithdrawTarget
	err := gdb.
		Where(&model.WithdrawTarget{WithdrawID: withdrawID}).
		First(&result).Error
	if err != nil {
		return model.WithdrawTarget{}, err
	}

	return result, nil
}

func GetLastWithdrawTargetByStatus(db database.Context, status model.WithdrawStatus) ([]model.WithdrawTarget, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}

	if len(status) == 0 {
		return nil, ErrInvalidWithdrawStatus
	}

	subQueryLast := gdb.Model(&model.WithdrawInfo{}).
		Select("MAX(id)").
		Group("withdraw_id").
		SubQuery()

	subQueryInfo := gdb.Model(&model.WithdrawInfo{}).
		Select("withdraw_id").
		Where("id IN (?)", subQueryLast).
		Where(model.WithdrawInfo{
			Status: status,
		}).
		SubQuery()

	var list []*model.WithdrawTarget
	err := gdb.Model(&model.WithdrawTarget{}).
		Joins("JOIN (?) AS i ON withdraw_target.withdraw_id = i.withdraw_id", subQueryInfo).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return converWithdrawTarget(list), nil
}

func converWithdrawTarget(list []*model.WithdrawTarget) []model.WithdrawTarget {
	var result []model.WithdrawTarget
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]

}
