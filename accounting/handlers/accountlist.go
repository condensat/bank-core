// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func AccountList(ctx context.Context, userID uint64) ([]common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountList")
	var result []common.AccountInfo

	log = log.WithField("UserID", userID)

	// Acquire Lock
	lock, err := cache.LockUser(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock user")
		return result, cache.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db bank.Database) error {
		accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, model.UserID(userID), "*", "*")
		if err != nil {
			return err
		}

		for _, account := range accounts {
			currency, err := database.GetCurrencyByName(db, account.CurrencyName)
			if err != nil {
				return err
			}
			accountState, err := database.GetAccountStatusByAccountID(db, account.ID)
			if err != nil {
				return err
			}

			last, err := database.GetLastAccountOperation(db, account.ID)
			if err != nil {
				return err
			}

			var balance float64
			var totalLocked float64
			if last.IsValid() {
				balance = float64(*last.Balance)
				totalLocked = float64(*last.TotalLocked)
			}

			result = append(result, common.AccountInfo{
				Timestamp: last.Timestamp,
				AccountID: uint64(account.ID),
				Currency: common.CurrencyInfo{
					Name:             string(currency.Name),
					Crypto:           currency.IsCrypto(),
					DisplayPrecision: uint(currency.DisplayPrecision()),
				},
				Name:        string(account.Name),
				Status:      string(accountState.State),
				Balance:     float64(balance),
				TotalLocked: float64(totalLocked),
			})
		}

		return nil
	})

	if err == nil {
		log.WithField("Count", len(result)).
			Debug("User accounts retrieved")
	}

	return result, err
}

func OnAccountList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.UserAccounts
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"UserID": request.UserID,
			})

			accounts, err := AccountList(ctx, request.UserID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to list user accounts")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.UserAccounts{
				UserID:   request.UserID,
				Accounts: accounts[:],
			}, nil
		})
}
