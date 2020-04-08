// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/jinzhu/gorm"
)

func GetLastCurencyRates(ctx context.Context) ([]model.CurrencyRate, error) {
	db := appcontext.Database(ctx)
	if db == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	gdb := db.DB().(*gorm.DB)

	subQuery := gdb.Model(&model.CurrencyRate{}).
		Select("MAX(id) as id, MAX(timestamp) AS last").
		Group("name").
		SubQuery()

	var list []*model.CurrencyRate
	err := gdb.Joins("RIGHT JOIN (?) AS t1 ON currency_rate.id = t1.id AND timestamp = t1.last", subQuery).
		Order("name ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var result []model.CurrencyRate
	for _, entry := range list {
		result = append(result, *entry)
	}

	return result, nil
}

func AppendCurencyRates(ctx context.Context, currencyRates []model.CurrencyRate) error {
	log := logger.Logger(ctx).WithField("Method", "database.AppendCurencyRates")
	db := appcontext.Database(ctx)
	if db == nil {
		return errors.New("Invalid appcontext.Database")
	}

	return db.Transaction(func(tx bank.Database) error {
		txdb := tx.DB().(*gorm.DB)
		if txdb == nil {
			return errors.New("Invalid tx Database")
		}

		var resultErr error
		for _, rate := range currencyRates {
			err := txdb.Create(&rate).Error
			if err != nil {
				log.WithError(err).Warning("Failed to add CurrencyRate")
				resultErr = err // return only last error
				continue        // continue to insert if possible
			}
		}

		return resultErr
	})
}
