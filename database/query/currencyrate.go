// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

func GetLastCurencyRates(db database.Context) ([]model.CurrencyRate, error) {
	if db == nil {
		return nil, database.ErrInvalidDatabase
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

func AppendCurencyRates(db database.Context, currencyRates []model.CurrencyRate) error {
	return db.Transaction(func(tx database.Context) error {
		txdb := tx.DB().(*gorm.DB)
		if txdb == nil {
			return errors.New("Invalid tx Database")
		}

		for _, rate := range currencyRates {
			err := txdb.Create(&rate).Error
			if err != nil {
				return err // Rollback all previous writes
			}
		}

		return nil
	})
}
