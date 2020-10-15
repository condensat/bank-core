// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

func CurrencyList(ctx context.Context) (common.CurrencyList, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CurrencyList")
	var result common.CurrencyList

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db database.Context) error {

		// list currencies
		list, err := query.ListAllCurrency(db)
		if err != nil {
			log.WithError(err).Error("Failed to ListAllCurrency")
			return err
		}

		for _, currency := range list {
			result.Currencies = append(result.Currencies, common.CurrencyInfo{
				Name:             string(currency.Name),
				DisplayName:      string(currency.DisplayName),
				Available:        currency.IsAvailable(),
				AutoCreate:       currency.AutoCreate,
				Type:             common.CurrencyType(currency.GetType()),
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

func OnCurrencyList(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Currencying.OnCurrencyList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CurrencyList
	return messaging.HandleRequest(ctx, appcontext.AppName(ctx), message, &request,
		func(ctx context.Context, _ messaging.BankObject) (messaging.BankObject, error) {

			response, err := CurrencyList(ctx)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CurrencyList")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}
