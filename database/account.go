// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
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

		var result model.Account
		err := db.
			Where(model.Account{
				UserID:       account.UserID,
				CurrencyName: account.CurrencyName,
			}).
			Assign(account).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.Account{}, ErrInvalidDatabase
	}
}
