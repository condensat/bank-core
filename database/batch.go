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
	ErrInvalidBatchID        = errors.New("Invalid BatchID")
	ErrInvalidBatchWithdraws = errors.New("Invalid Withdraws")
	ErrInvalidNetwork        = errors.New("Invalid Network")
)

func AddBatch(db bank.Database, network model.BatchNetwork, data model.BatchData, withdraws ...model.WithdrawID) (model.Batch, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Batch{}, errors.New("Invalid appcontext.Database")
	}

	if len(network) == 0 {
		return model.Batch{}, ErrInvalidNetwork
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.Batch{
		Timestamp: timestamp,
		Network:   network,
		Data:      data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.Batch{}, err
	}

	err = AddWithdrawToBatch(db, result.ID, withdraws...)
	if err != nil {
		return model.Batch{}, err
	}

	return result, nil
}

func GetBatch(db bank.Database, ID model.BatchID) (model.Batch, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Batch{}, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return model.Batch{}, ErrInvalidBatchID
	}

	var result model.Batch
	err := gdb.
		Where(&model.Batch{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.Batch{}, err
	}

	return result, nil
}

func ListBatchNetworksByStatus(db bank.Database, status model.BatchStatus) ([]model.BatchNetwork, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if len(status) == 0 {
		return nil, ErrInvalidWithdrawStatus
	}

	subQueryInfo := gdb.Model(&model.BatchInfo{}).
		Where(model.BatchInfo{
			Status: status,
		}).
		SubQuery()

	var list []*model.Batch
	err := gdb.Model(&model.Batch{}).
		Select("network").
		Joins("JOIN (?) AS i ON batch.id = i.batch_id", subQueryInfo).
		Group("network").
		Order("batch.id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertBatchNetworkList(list), nil
}

func convertBatchNetworkList(list []*model.Batch) []model.BatchNetwork {
	var result []model.BatchNetwork
	for _, curr := range list {
		if curr != nil {
			result = append(result, curr.Network)
		}
	}

	return result[:]
}
