// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/accounting/internal"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencyList(ctx context.Context) (common.CurrencyList, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CurrencyList")
	var result common.CurrencyList

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		// list currencies
		list, err := database.ListAllCurrency(db)
		if err != nil {
			log.WithError(err).Error("Failed to ListAllCurrency")
			return err
		}

		for _, currency := range list {
			result.Currencies = append(result.Currencies, common.CurrencyInfo{
				Name:             string(currency.Name),
				Available:        currency.IsAvailable(),
				Crypto:           currency.IsCrypto(),
				DisplayPrecision: uint(currency.DisplayPrecision()),
			})
		}

		return nil
	})

	if err == nil {
		log.WithFields(logrus.Fields{
			"Count": len(result.Currencies),
		}).Trace("Currency list")
	}

	return result, err
}

func OnCurrencyList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Currencying.OnCurrencyList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CurrencyList
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {

			response, err := CurrencyList(ctx)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CurrencyList")
				return nil, internal.ErrInternalError
			}

			// return response
			return &response, nil
		})
}
