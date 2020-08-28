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
	ErrInvalidFeeID     = errors.New("Invalid FeeID")
	ErrInvalidFeeAmount = errors.New("Invalid Fee Amount")
)

func AddFee(db bank.Database, withdrawID model.WithdrawID, amount model.Float, data model.FeeData) (model.Fee, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Fee{}, errors.New("Invalid appcontext.Database")
	}

	if withdrawID == 0 {
		return model.Fee{}, ErrInvalidWithdrawID
	}
	if amount <= 0.0 {
		return model.Fee{}, ErrInvalidFeeAmount
	}

	result := model.Fee{
		WithdrawID: withdrawID,
		Amount:     &amount,
		Data:       data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.Fee{}, err
	}

	return result, nil
}

func GetFee(db bank.Database, ID model.FeeID) (model.Fee, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Fee{}, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return model.Fee{}, ErrInvalidFeeID
	}

	var result model.Fee
	err := gdb.
		Where(&model.Fee{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.Fee{}, err
	}

	return result, nil
}

func GetFeeByWithdrawID(db bank.Database, withdrawID model.WithdrawID) (model.Fee, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Fee{}, errors.New("Invalid appcontext.Database")
	}

	if withdrawID == 0 {
		return model.Fee{}, ErrInvalidWithdrawID
	}

	var result model.Fee
	err := gdb.
		Where(&model.Fee{WithdrawID: withdrawID}).
		First(&result).Error
	if err != nil {
		return model.Fee{}, err
	}

	return result, nil
}
