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
	ErrInvalidOperationAmount          = errors.New("Invalid Operation Amount")
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
	if operation.Amount < 0.0 {
		return model.OperationInfo{}, ErrInvalidOperationAmount
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
func GetOperationInfo(db bank.Database, operationID model.OperationInfoID) (model.OperationInfo, error) {
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
func GetOperationInfoByCryptoAddress(db bank.Database, cryptoAddressID model.CryptoAddressID) ([]model.OperationInfo, error) {
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

func FindCryptoAddressesNotInOperationInfo(db bank.Database, chain model.String) ([]model.CryptoAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	/*
		select l.* from crypto_address as l
		WHERE l.chain = 'bitcoin-testnet'
		  AND l.id NOT IN
		(
			SELECT  r.crypto_address_id
			FROM    operation_info as r
		);
	*/

	var list []*model.CryptoAddress
	notInQuery := gdb.Model(&model.OperationInfo{}).
		Select("crypto_address_id").
		SubQuery()

	err := gdb.Model(&model.CryptoAddress{}).
		Where("chain = ?", chain).
		Where("id NOT IN ?", notInQuery).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return converCryptoAddressList(list), nil
}

func FindCryptoAddressesByOperationInfoState(db bank.Database, chain model.String, states ...model.String) ([]model.CryptoAddress, error) {
	gdb := db.DB().(*gorm.DB)

	uniq := make(map[model.String]model.String)
	for _, state := range states {
		// skip empty states
		if len(state) == 0 {
			continue
		}
		uniq[state] = state
	}

	// slice uniq map
	var slice []model.String
	for state := range uniq {
		slice = append(slice, state)
	}

	// state must not be empty
	if len(slice) == 0 {
		return nil, ErrInvalidOperationStatus
	}

	subQueryState := gdb.Model(&model.OperationStatus{}).
		Select("operation_info_id").
		Where("state IN (?)", slice).
		SubQuery()
	subQueryInfo := gdb.Model(&model.OperationInfo{}).
		Select("id, crypto_address_id").
		SubQuery()

	var list []*model.CryptoAddress
	err := gdb.Model(&model.CryptoAddress{}).
		Joins("JOIN (?) AS oi ON crypto_address.id = oi.crypto_address_id", subQueryInfo).
		Joins("JOIN (?) AS os ON oi.id = os.operation_info_id", subQueryState).
		Where("chain = ?", chain).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return converCryptoAddressList(list), nil
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
