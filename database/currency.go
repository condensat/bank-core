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
	FlagCurencyAll = 0

	FlagCurencyDisable   = 0
	FlagCurencyAvailable = 1
)

var (
	ErrInvalidCurrencyName  = errors.New("Invalid Currency Name")
	ErrCurrencyNotFound     = errors.New("Currency Not found")
	ErrCurrencyNotAvailable = errors.New("Currency Not Available")
	ErrCurrencyNotCrypto    = errors.New("Currency Not Crypto")
)

// AddOrUpdateCurrency
func AddOrUpdateCurrency(db bank.Database, currency model.Currency) (model.Currency, error) {
	var result model.Currency
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	err := gdb.
		Where(model.Currency{
			Name: currency.Name,
		}).
		Assign(currency).
		FirstOrCreate(&result).Error

	return result, err
}

// CurrencyExists
func CurrencyExists(db bank.Database, name model.CurrencyName) bool {
	entry, err := GetCurrencyByName(db, name)

	return err == nil && entry.Name == name
}

// GetCurrencyByName
func GetCurrencyByName(db bank.Database, name model.CurrencyName) (model.Currency, error) {
	var result model.Currency

	list, err := QueryCurrencyList(db, name, FlagCurencyAll)
	if len(list) > 0 {
		result = list[0]
	}

	return result, err
}

// CountCurrencies
func CountCurrencies(db bank.Database) int {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return 0
	}

	var count int
	gdb.Model(&model.Currency{}).Count(&count)
	return count
}

// ListAllCurrency
func ListAllCurrency(db bank.Database) ([]model.Currency, error) {
	return QueryCurrencyList(db, "", FlagCurencyAll)
}

// ListAvailableCurrency
func ListAvailableCurrency(db bank.Database) ([]model.Currency, error) {
	return QueryCurrencyList(db, "", FlagCurencyAvailable)
}

// QueryCurrencyList
func QueryCurrencyList(db bank.Database, name model.CurrencyName, available int) ([]model.Currency, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	var filters []func(db *gorm.DB) *gorm.DB
	if len(name) > 0 {
		filters = append(filters, ScopeCurencyName(name))
	}
	if available > 0 {
		filters = append(filters, ScopeCurencyAvailable(available))
	}

	var list []*model.Currency
	err := gdb.Model(&model.Currency{}).
		Scopes(filters...).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertCurrencyList(list), nil
}

// ScopeCurencyName
func ScopeCurencyName(name model.CurrencyName) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqName(), name)
	}
}

// ScopeCurencyAvailable
func ScopeCurencyAvailable(available int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqAvailable(), available)
	}
}

func convertCurrencyList(list []*model.Currency) []model.Currency {
	var result []model.Currency
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *curr)
	}

	return result[:]
}

const (
	colCurrencyName        = "name"
	colCurrencyDisplayName = "display_name"
	colCurrencyType        = "type"
	colCurrencyAvailable   = "available"
	colCurrencyCrypto      = "crypto"
	colCurrencyPrecision   = "precision"
	colCurrencyAutoCreate  = "auto_create"
)

func currencyColumnNames() []string {
	return []string{
		colCurrencyName,
		colCurrencyDisplayName,
		colCurrencyType,
		colCurrencyAvailable,
		colCurrencyCrypto,
		colCurrencyPrecision,
		colCurrencyAutoCreate,
	}
}

// zero allocation requests string for scope
func reqName() string {
	var req [len(colCurrencyName) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colCurrencyName)
	copy(req[off:], reqEQ)

	return string(req[:])
}

func reqAvailable() string {
	var req [len(colCurrencyAvailable) + len(reqGTE)]byte
	off := 0
	off += copy(req[off:], colCurrencyAvailable)
	copy(req[off:], reqGTE)

	return string(req[:])
}
