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

const (
	AccountNameDefault  = "default"
	AccountNameWildcard = "*"
)

var (
	ErrAccountExists = errors.New("Account Exists")
)

func CreateAccount(db bank.Database, account model.Account) (model.Account, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if len(account.Name) == 0 {
			account.Name = AccountNameDefault
		}

		if !UserExists(db, account.UserID) {
			return model.Account{}, ErrUserNotFound
		}

		if !CurrencyExists(db, account.CurrencyName) {
			return model.Account{}, ErrCurrencyNotFound
		}

		if AccountsExists(db, account.UserID, account.CurrencyName, account.Name) {
			return model.Account{}, ErrAccountExists
		}

		var result model.Account
		err := gdb.
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
func AccountsExists(db bank.Database, userID uint64, currency, name string) bool {
	entries, err := GetAccountsByUserAndCurrencyAndName(db, userID, currency, name)

	return err == nil && len(entries) > 0
}

// GetAccountsByNameAndCurrency
func GetAccountsByUserAndCurrencyAndName(db bank.Database, userID uint64, currency, name string) ([]model.Account, error) {
	return QueryAccountList(db, userID, currency, name)
}

// QueryAccountList
func QueryAccountList(db bank.Database, userID uint64, currency, name string) ([]model.Account, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	var filters []func(db *gorm.DB) *gorm.DB
	if userID == 0 {
		return nil, errors.New("UserId is mandatory")
	}

	// default account name if empty
	if len(name) == 0 {
		name = AccountNameDefault
	}

	filters = append(filters, ScopeUserID(userID))
	filters = append(filters, ScopeAccountCurrencyName(currency))
	// manage wildcard search (no account name filter selects all accounts)
	if name != "*" {
		filters = append(filters, ScopeAccountName(name))
	}

	var list []*model.Account
	err := gdb.Model(&model.Account{}).
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
	colID                  = "id"
	colUserID              = "user_id"
	colAccountCurrencyName = "currency_name"
	colAccountName         = "name"
)

func accountColumnNames() []string {
	return []string{
		colID,
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
