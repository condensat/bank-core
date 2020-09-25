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
	ErrAccountExists   = errors.New("Account Exists")
	ErrAccountNotFound = errors.New("Account Not Found")
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

		if err != nil {
			return model.Account{}, err
		}

		// Create init operation
		_, err = txApppendAccountOperation(db, model.NewInitOperation(result.ID, 0))
		if err != nil {
			return model.Account{}, err
		}

		return result, err

	default:
		return model.Account{}, ErrInvalidDatabase
	}
}

// AccountsExists
func AccountsExists(db bank.Database, userID model.UserID, currency model.CurrencyName, name model.AccountName) bool {
	entries, err := GetAccountsByUserAndCurrencyAndName(db, userID, currency, name)

	return err == nil && len(entries) > 0
}

func GetAccountByID(db bank.Database, accountID model.AccountID) (model.Account, error) {
	var result model.Account

	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	err := gdb.Model(&model.Account{}).
		Scopes(ScopeAccountID(accountID)).
		First(&result).Error

	return result, err
}

// GetAccountsByNameAndCurrency
func GetAccountsByUserAndCurrencyAndName(db bank.Database, userID model.UserID, currency model.CurrencyName, name model.AccountName) ([]model.Account, error) {
	return QueryAccountList(db, userID, currency, name)
}

type AccountSummary struct {
	CurrencyName string
	Balance      float64
	TotalLocked  float64
}

type AccountInfos struct {
	Count    int
	Active   int
	Accounts []AccountSummary
}

func AccountsInfos(db bank.Database) (AccountInfos, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return AccountInfos{}, errors.New("Invalid appcontext.Database")
	}

	var totalAccounts int64
	err := gdb.Model(&model.Account{}).
		Count(&totalAccounts).Error
	if err != nil {
		return AccountInfos{}, err
	}

	var activeAccounts int64
	err = gdb.Model(&model.AccountState{}).
		Where(&model.AccountState{
			State: model.AccountStatusNormal,
		}).Count(&activeAccounts).Error
	if err != nil {
		return AccountInfos{}, err
	}

	subQueryLast := gdb.Model(&model.AccountOperation{}).
		Select("MAX(id)").
		Group("account_id").
		SubQuery()

	subQueryAccount := gdb.Model(&model.Account{}).
		Select("id as aid, currency_name").
		SubQuery()

	var list []*AccountSummary
	err = gdb.Table("account_operation").
		Joins("JOIN (?) AS a ON a.aid = account_id", subQueryAccount).
		Where("id IN (?)", subQueryLast).
		Group("currency_name").
		Select("currency_name, SUM(balance) as balance, SUM(total_locked) as total_locked").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return AccountInfos{}, err
	}

	return AccountInfos{
		Count:    int(totalAccounts),
		Active:   int(activeAccounts),
		Accounts: convertAccountSummaryList(list),
	}, nil
}

func convertAccountSummaryList(list []*AccountSummary) []AccountSummary {
	var result []AccountSummary
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}

// QueryAccountList
func QueryAccountList(db bank.Database, userID model.UserID, currency model.CurrencyName, name model.AccountName) ([]model.Account, error) {
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
	// manage wildcards
	if currency != "*" {
		filters = append(filters, ScopeAccountCurrencyName(currency))
	}
	if name != "*" {
		filters = append(filters, ScopeAccountName(name))
	}

	var list []*model.Account
	err := gdb.Model(&model.Account{}).
		Scopes(filters...).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertAccountList(list), nil
}

// ScopeAccountID
func ScopeAccountID(accountID model.AccountID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqAccountID(), accountID)
	}
}

// ScopeUserID
func ScopeUserID(userID model.UserID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqUserID(), userID)
	}
}

// ScopeCurencyName
func ScopeAccountCurrencyName(name model.CurrencyName) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqAccountCurrencyName(), name)
	}
}

// ScopeAccountName
func ScopeAccountName(name model.AccountName) func(db *gorm.DB) *gorm.DB {
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
func reqAccountID() string {
	var req [len(colID) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colID)
	copy(req[off:], reqEQ)

	return string(req[:])
}

// zero allocation requests string for scope
func reqUserID() string {
	var req [len(colUserID) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colUserID)
	copy(req[off:], reqEQ)

	return string(req[:])
}

func reqAccountCurrencyName() string {
	var req [len(colAccountCurrencyName) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colAccountCurrencyName)
	copy(req[off:], reqEQ)

	return string(req[:])
}

func reqAccountName() string {
	var req [len(colAccountName) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colAccountName)
	copy(req[off:], reqEQ)

	return string(req[:])
}
