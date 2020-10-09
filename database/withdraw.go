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

func FindWithdrawByCurrencyNameAndStatus(db bank.Database, currency model.CurrencyName, status model.WithdrawStatus) ([]model.Withdraw, error) {
	gdb := db.DB().(*gorm.DB)

	subQueryAccount := gdb.Model(&model.Account{}).
		Select("id").
		Where("currency_name = ?", currency).
		SubQuery()

	subQueryInfo := gdb.Model(&model.WithdrawInfo{}).
		Select("withdraw_id").
		Where("status = ?", status).
		SubQuery()

	var list []*model.Withdraw
	err := gdb.Model(&model.Withdraw{}).
		Joins("JOIN (?) AS a ON withdraw.from = a.id", subQueryAccount).
		Joins("JOIN (?) AS i ON withdraw.id = i.withdraw_id", subQueryInfo).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertWithdraws(list), nil
}

func FindWithdrawByUser(db bank.Database, userID model.UserID) ([]model.Withdraw, error) {
	gdb := db.DB().(*gorm.DB)

	if userID == 0 {
		return nil, ErrInvalidUserID
	}

	subQueryAccount := gdb.Model(&model.Account{}).
		Select("id").
		Where("user_id = ?", userID).
		SubQuery()

	var list []*model.Withdraw
	err := gdb.Model(&model.Withdraw{}).
		Joins("JOIN (?) AS a ON withdraw.from = a.id", subQueryAccount).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertWithdraws(list), nil
}

type WithdrawInfos struct {
	Count  int
	Active int
}

func WithdrawsInfos(db bank.Database) (WithdrawInfos, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return WithdrawInfos{}, errors.New("Invalid appcontext.Database")
	}

	var totalWithdraws int64
	err := gdb.Model(&model.Withdraw{}).
		Count(&totalWithdraws).Error
	if err != nil {
		return WithdrawInfos{}, err
	}

	subQueryLast := gdb.Model(&model.WithdrawInfo{}).
		Select("MAX(id)").
		Group("withdraw_id").
		SubQuery()

	var activeWithdraws int64
	err = gdb.Model(&model.WithdrawInfo{}).
		Where("id IN (?)", subQueryLast).
		Where("status <> ?", model.WithdrawStatusSettled).
		Count(&activeWithdraws).Error
	if err != nil {
		return WithdrawInfos{}, err
	}

	return WithdrawInfos{
		Count:  int(totalWithdraws),
		Active: int(activeWithdraws),
	}, nil
}

func convertWithdraws(list []*model.Withdraw) []model.Withdraw {
	var result []model.Withdraw
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}
