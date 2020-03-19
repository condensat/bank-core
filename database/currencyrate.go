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

func AppendCurencyRates(ctx context.Context, currencyRates []model.CurrencyRate) error {
	log := logger.Logger(ctx).WithField("Method", "database.AppendCurencyRates")
	db := appcontext.Database(ctx)
	if db == nil {
		return errors.New("Invalid appcontext.Database")
	}

	return db.Transaction(func(tx bank.Database) error {
		txdb := tx.DB().(*gorm.DB)
		if db == nil {
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
