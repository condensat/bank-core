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
	ErrFeeInfoInvalid = errors.New("Invalid FeeInfo")
)

// AddOrUpdateFeeInfo
func AddOrUpdateFeeInfo(db bank.Database, feeInfo model.FeeInfo) (model.FeeInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.FeeInfo{}, errors.New("Invalid appcontext.Database")
	}

	if !feeInfo.IsValid() {
		return model.FeeInfo{}, ErrFeeInfoInvalid
	}

	var result model.FeeInfo
	err := gdb.
		Where(model.FeeInfo{
			Currency: feeInfo.Currency,
		}).
		Assign(feeInfo).
		FirstOrCreate(&result).Error

	return result, err
}

// FeeInfoExists
func FeeInfoExists(db bank.Database, currency model.CurrencyName) bool {
	entry, err := GetFeeInfo(db, currency)

	return err == nil && entry.Currency == currency && entry.IsValid()
}

// GetFeeInfo
func GetFeeInfo(db bank.Database, currency model.CurrencyName) (model.FeeInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.FeeInfo{}, errors.New("Invalid appcontext.Database")
	}

	if len(currency) == 0 {
		return model.FeeInfo{}, ErrInvalidCurrencyName
	}

	var result model.FeeInfo
	err := gdb.
		Where(&model.FeeInfo{
			Currency: currency,
		}).First(&result).Error
	if err != nil {
		return model.FeeInfo{}, err
	}

	if result.Currency != currency || !result.IsValid() {
		return model.FeeInfo{}, ErrFeeInfoInvalid
	}

	return result, nil
}
