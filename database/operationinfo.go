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
	ErrInvalidOperationInfoID          = errors.New("Invalid OperationInfo")
	ErrOperationInfoUpdateNotPermitted = errors.New("OperationInfo Update Not Permitted")
	ErrInvalidTransactionID            = errors.New("Invalid Transaction ID")
)

// AddOperationInfo
func AddOperationInfo(db bank.Database, operation model.OperationInfo) (model.OperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.OperationInfo{}, errors.New("Invalid appcontext.Database")
	}

	if operation.ID != 0 {
		return model.OperationInfo{}, ErrOperationInfoUpdateNotPermitted
	}

	if operation.CryptoAddressID == 0 {
		return model.OperationInfo{}, ErrInvalidCryptoAddressID
	}

	if len(operation.TxID) == 0 {
		return model.OperationInfo{}, ErrInvalidTransactionID
	}

	operation.Timestamp = time.Now().UTC().Truncate(time.Second)

	err := gdb.
		Assign(operation).
		Create(&operation).Error
	if err != nil {
		return model.OperationInfo{}, err
	}

	return operation, nil
}

// GetOperationInfo
func GetOperationInfo(db bank.Database, operationID model.ID) (model.OperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.OperationInfo{}, errors.New("Invalid appcontext.Database")
	}

	if operationID == 0 {
		return model.OperationInfo{}, ErrInvalidOperationInfoID
	}

	var result model.OperationInfo
	err := gdb.
		Where(model.OperationInfo{
			ID: operationID,
		}).
		First(&result).Error
	if err != nil {
		return model.OperationInfo{}, err
	}

	return result, nil
}

// GetOperationInfoByTxId
func GetOperationInfoByTxId(db bank.Database, txID model.TxID) (model.OperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.OperationInfo{}, errors.New("Invalid appcontext.Database")
	}

	if len(txID) == 0 {
		return model.OperationInfo{}, ErrInvalidTransactionID
	}

	var result model.OperationInfo
	err := gdb.
		Where(model.OperationInfo{
			TxID: txID,
		}).
		First(&result).Error
	if err != nil {
		return model.OperationInfo{}, err
	}

	return result, nil
}

// GetOperationInfoByCryptoAddress
func GetOperationInfoByCryptoAddress(db bank.Database, cryptoAddressID model.ID) ([]model.OperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if cryptoAddressID == 0 {
		return nil, ErrInvalidCryptoAddressID
	}

	var list []*model.OperationInfo
	err := gdb.
		Where(model.OperationInfo{
			CryptoAddressID: cryptoAddressID,
		}).
		Find(&list).Error
	if err != nil {
		return nil, err
	}

	return convertOperationInfoList(list), nil
}

func convertOperationInfoList(list []*model.OperationInfo) []model.OperationInfo {
	var result []model.OperationInfo
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *curr)
	}

	return result[:]
}
