// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/sirupsen/logrus"
)

func AccountInfo(ctx context.Context, accountID uint64) (common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountInfo")

	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
	})

	var result common.AccountInfo
	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		account, err := database.GetAccountByID(db, model.AccountID(accountID))
		if err != nil {
			return err
		}

		result, err = txGetAccountInfo(db, account)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.WithError(err).
			Error("Failed to get AccountInfo")
		return common.AccountInfo{}, err
	}

	return result, nil
}

func txGetAccountInfo(db bank.Database, account model.Account) (common.AccountInfo, error) {
	currency, err := database.GetCurrencyByName(db, account.CurrencyName)
	if err != nil {
		return common.AccountInfo{}, err
	}
	accountState, err := database.GetAccountStatusByAccountID(db, account.ID)
	if err != nil {
		return common.AccountInfo{}, err
	}

	last, err := database.GetLastAccountOperation(db, account.ID)
	if err != nil {
		return common.AccountInfo{}, err
	}

	var balance float64
	var totalLocked float64
	if last.IsValid() {
		balance = float64(*last.Balance)
		totalLocked = float64(*last.TotalLocked)
	}

	return common.AccountInfo{
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
	}, nil
}

func OnAccountInfo(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountInfo")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
			})

			info, err := AccountInfo(ctx, request.AccountID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to get AccountInfo")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &info, nil
		})
}
