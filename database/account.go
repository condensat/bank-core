// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"errors"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrAccountExists = errors.New("Account Exists")
)

func CreateAccount(ctx context.Context, account model.Account) (model.Account, error) {
	db := appcontext.Database(ctx)
	switch db := db.DB().(type) {
	case *gorm.DB:

		if !UserExists(ctx, account.UserID) {
			return model.Account{}, ErrUserNotFound
		}

		if !CurrencyExists(ctx, account.CurrencyName) {
			return model.Account{}, ErrCurrencyNotFound
		}

		if AccountsExists(ctx, account.UserID, account.CurrencyName, account.Name) {
			return model.Account{}, ErrAccountExists
		}

		var result model.Account
		err := db.
			Where(model.Account{
				UserID:       account.UserID,
				CurrencyName: account.CurrencyName,
				Name:         account.Name,
			}).
			Assign(account).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.Account{}, ErrInvalidDatabase
	}
}

// AccountsExists
func AccountsExists(ctx context.Context, userID uint64, currency, name string) bool {
	entries, err := GetAccountsByUserAndCurrencyAndName(ctx, userID, currency, name)

	return err == nil && len(entries) > 0
}

// GetAccountsByNameAndCurrency
func GetAccountsByUserAndCurrencyAndName(ctx context.Context, userID uint64, currency, name string) ([]model.Account, error) {
	return QueryAccountList(ctx, userID, currency, name)
}

// QueryAccountList
func QueryAccountList(ctx context.Context, userID uint64, currency, name string) ([]model.Account, error) {
	db := appcontext.Database(ctx).DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	var filters []func(db *gorm.DB) *gorm.DB
	if userID == 0 {
		return nil, errors.New("UserId is mandatory")
	}

	filters = append(filters, ScopeUserID(userID))
	if len(currency) > 0 {
		filters = append(filters, ScopeAccountCurrencyName(currency))
	}
	if len(currency) > 0 {
		filters = append(filters, ScopeAccountCurrencyName(currency))
	}
	if len(name) > 0 {
		filters = append(filters, ScopeAccountName(name))
	}

	var list []*model.Account
	err := db.Model(&model.Account{}).
		Scopes(filters...).
		Find(&list).Error

	return convertAccountList(list), err
}

// ScopeUserID
func ScopeUserID(userID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqUserID(), userID)
	}
}

// ScopeCurencyName
func ScopeAccountCurrencyName(name string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqAccountCurrencyName(), name)
	}
}

// ScopeAccountName
func ScopeAccountName(name string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqAccountName(), name)
	}
}

func convertAccountList(list []*model.Account) []model.Account {
	var result []model.Account
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *curr)
	}

	return result[:]
}

const (
	colUserID              = "user_id"
	colAccountCurrencyName = "currency_name"
	colAccountName         = "name"
)

func currencyUserID() []string {
	return []string{
		colUserID,
		colAccountCurrencyName,
		colAccountName,
	}
}

// zero allocation requests string for scope
func reqUserID() string {
	var req [len(colUserID) + len(reqGTE)]byte
	off := 0
	off += copy(req[off:], colUserID)
	off += copy(req[off:], reqGTE)

	return string(req[:])
}

func reqAccountCurrencyName() string {
	var req [len(colAccountCurrencyName) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colAccountCurrencyName)
	off += copy(req[off:], reqEQ)

	return string(req[:])
}

func reqAccountName() string {
	var req [len(colAccountName) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colAccountName)
	off += copy(req[off:], reqEQ)

	return string(req[:])
}
