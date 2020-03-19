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

const (
	FlagCurencyAll       = 0
	FlagCurencyAvailable = 1
)

// AddOrUpdateCurrency
func AddOrUpdateCurrency(ctx context.Context, currency model.Currency) (model.Currency, error) {
	var result model.Currency
	db := appcontext.Database(ctx).DB().(*gorm.DB)
	if db == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	err := db.
		Where(model.Currency{
			Name: currency.Name,
		}).
		Assign(currency).
		FirstOrCreate(&result).Error

	return result, err
}

// GetCurrencyByName
func GetCurrencyByName(ctx context.Context, name string) (model.Currency, error) {
	var result model.Currency

	list, err := QueryCurrencyList(ctx, name, FlagCurencyAll)
	if len(list) > 0 {
		result = list[0]
	}

	return result, err
}

// CountCurrencies
func CountCurrencies(ctx context.Context) int {
	db := appcontext.Database(ctx).DB().(*gorm.DB)
	if db == nil {
		return 0
	}

	var count int
	db.Model(&model.Currency{}).Count(&count)
	return count
}

// ListAllCurrency
func ListAllCurrency(ctx context.Context) ([]model.Currency, error) {
	return QueryCurrencyList(ctx, "", FlagCurencyAll)
}

// ListAvailableCurrency
func ListAvailableCurrency(ctx context.Context) ([]model.Currency, error) {
	return QueryCurrencyList(ctx, "", FlagCurencyAvailable)
}

// QueryCurrencyList
func QueryCurrencyList(ctx context.Context, name string, available int) ([]model.Currency, error) {
	db := appcontext.Database(ctx).DB().(*gorm.DB)
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
	err := db.Model(&model.Currency{}).
		Scopes(filters...).
		Find(&list).Error

	return convertCurrencyList(list), err
}

// ScopeCurencyName
func ScopeCurencyName(name string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	}
}

// ScopeCurencyAvailable
func ScopeCurencyAvailable(available int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("available >= ?", available)
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
