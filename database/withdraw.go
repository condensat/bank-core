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
	ErrInvalidWithdrawID              = errors.New("Invalid WithdrawID")
	ErrInvalidWithdrawAmount          = errors.New("Invalid Amount")
	ErrInvalidWithdrawAccount         = errors.New("Invalid Withdraw Account")
	ErrInvalidWithdrawAccountCurrency = errors.New("Invalid Withdraw Account Currency")
	ErrInvalidBatchMode               = errors.New("Invalid BatchMode")
)

func AddWithdraw(db bank.Database, from, to model.AccountID, amount model.Float, batch model.BatchMode, data model.WithdrawData) (model.Withdraw, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Withdraw{}, errors.New("Invalid appcontext.Database")
	}

	if from == 0 {
		return model.Withdraw{}, ErrInvalidAccountID
	}
	if to == 0 {
		return model.Withdraw{}, ErrInvalidAccountID
	}
	if from == to {
		return model.Withdraw{}, ErrInvalidAccountID
	}
	accountFrom, err := GetAccountByID(db, from)
	if err != nil {
		return model.Withdraw{}, ErrInvalidWithdrawAccount
	}
	accountTo, err := GetAccountByID(db, to)
	if err != nil {
		return model.Withdraw{}, ErrInvalidWithdrawAccount
	}
	if accountFrom.CurrencyName != accountTo.CurrencyName {
		return model.Withdraw{}, ErrInvalidWithdrawAccountCurrency
	}

	if amount <= 0.0 {
		return model.Withdraw{}, ErrInvalidWithdrawAmount
	}
	if len(batch) == 0 {
		return model.Withdraw{}, ErrInvalidBatchMode
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.Withdraw{
		Timestamp: timestamp,
		From:      from,
		To:        to,
		Amount:    &amount,
		Batch:     batch,
		Data:      data,
	}
	err = gdb.Create(&result).Error
	if err != nil {
		return model.Withdraw{}, err
	}

	return result, nil
}

func GetWithdraw(db bank.Database, ID model.WithdrawID) (model.Withdraw, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Withdraw{}, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return model.Withdraw{}, ErrInvalidWithdrawID
	}

	var result model.Withdraw
	err := gdb.
		Where(&model.Withdraw{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.Withdraw{}, err
	}

	return result, nil
}
