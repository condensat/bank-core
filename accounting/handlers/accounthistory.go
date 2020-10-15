// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func AccountHistory(ctx context.Context, accountID uint64, from, to time.Time) (string, string, []common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountHistory")

	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
		"From":      from,
		"To":        to,
	})

	// Database Query
	db := appcontext.Database(ctx)
	account, err := query.GetAccountByID(db, model.AccountID(accountID))
	if err != nil {
		return "", "", nil, err
	}
	currency, err := query.GetCurrencyByName(db, account.CurrencyName)
	if err != nil {
		return "", "", nil, err
	}

	isAsset := strings.HasPrefix(string(currency.Name), "Li#")
	tickerPrecision := -1 // no ticker precison
	if isAsset {
		tickerPrecision = 0
	}

	operations, err := query.GeAccountHistoryRange(db, account.ID, from, to)
	if err != nil {
		return "", "", nil, err
	}

	var result []common.AccountEntry
	for _, op := range operations {
		if !op.IsValid() {
			log.WithError(query.ErrInvalidAccountOperation).
				Warn("Invalid operation in history")
			continue
		}

		result = append(result, common.AccountEntry{
			OperationID: uint64(op.ID),

			AccountID: uint64(op.AccountID),
			Currency:  string(account.CurrencyName),

			OperationType:    string(op.OperationType),
			SynchroneousType: string(op.SynchroneousType),

			Timestamp: op.Timestamp,
			Label:     "N/A",
			Amount:    convertAssetAmount(float64(*op.Amount), tickerPrecision),
			Balance:   convertAssetAmount(float64(*op.Balance), tickerPrecision),

			LockAmount:  convertAssetAmount(float64(*op.LockAmount), tickerPrecision),
			TotalLocked: convertAssetAmount(float64(*op.TotalLocked), tickerPrecision),
		})
	}

	log.
		WithField("Count", len(result)).
		Debug("Account history retrieved")

	return string(account.CurrencyName), string(currency.DisplayName), result, nil
}

func OnAccountHistory(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountHistory")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountHistory
	return messaging.HandleRequest(ctx, appcontext.AppName(ctx), message, &request,
		func(ctx context.Context, _ messaging.BankObject) (messaging.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
			})

			currency, displayName, entries, err := AccountHistory(ctx, request.AccountID, request.From, request.To)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to get AccountHistory")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.AccountHistory{
				AccountID:   request.AccountID,
				DisplayName: displayName,
				Ticker:      currency,
				From:        request.From,
				To:          request.To,

				Entries: entries,
			}, nil
		})
}
