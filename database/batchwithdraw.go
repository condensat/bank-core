// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidBatchWithdrawID = errors.New("Invalid BatchWithdrawID")
)

func AddWithdrawToBatch(db bank.Database, batchID model.BatchID, withdraws ...model.WithdrawID) error {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return errors.New("Invalid appcontext.Database")
	}

	if batchID == 0 {
		return ErrInvalidBatchID
	}

	for _, wID := range withdraws {
		if wID == 0 {
			return ErrInvalidBatchWithdrawID
		}
	}

	for _, wID := range withdraws {
		entry := model.BatchWithdraw{
			BatchID:    batchID,
			WithdrawID: wID,
		}
		err := gdb.Create(&entry).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func GetBatchWithdraws(db bank.Database, batchID model.BatchID) ([]model.WithdrawID, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if batchID == 0 {
		return nil, ErrInvalidBatchID
	}

	var list []*model.BatchWithdraw
	err := gdb.
		Where(model.BatchWithdraw{
			BatchID: batchID,
		}).
		Order("batch_id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertWithdrawList(list), nil
}

func convertWithdrawList(list []*model.BatchWithdraw) []model.WithdrawID {
	var result []model.WithdrawID
	for _, curr := range list {
		if curr != nil {
			result = append(result, curr.WithdrawID)
		}
	}

	return result[:]
}
