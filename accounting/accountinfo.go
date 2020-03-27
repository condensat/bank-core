// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
)

func ListUserAccounts(ctx context.Context, userID uint64) ([]AccountInfo, error) {
	var result []AccountInfo

	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {
		accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, model.UserID(userID), "", "*")
		if err != nil {
			return err
		}

		for _, account := range accounts {
			accountState, err := database.GetAccountStatusByAccountID(db, account.ID)
			if err != nil {
				return err
			}

			result = append(result, AccountInfo{
				AccountID: uint64(account.ID),
				Currency:  string(account.CurrencyName),
				Name:      string(account.Name),
				Status:    string(accountState.State),
			})
		}

		return nil
	})

	return result, err
}

func GetAccountHistory(ctx context.Context, accountID uint64, from, to time.Time) ([]AccountEntry, error) {
	var result []AccountEntry
	return result, nil
}
