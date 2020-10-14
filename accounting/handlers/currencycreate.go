// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

func CurrencyCreate(ctx context.Context, currencyName, displayName string, currencyType common.CurrencyType, isCrypto bool, precision uint) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CurrencyCreate")
	var result common.CurrencyInfo

	log = log.WithField("CurrencyName", currencyName)

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db database.Context) error {

		// check if currency exists
		currency, err := query.GetCurrencyByName(db, model.CurrencyName(currencyName))
		if err != nil {
			log.WithError(err).Error("Failed to GetCurrencyByName")
			return err
		}

		// create if not exists
		if len(currency.Name) == 0 {
			var crypto int
			if isCrypto {
				crypto = 1
			}
			currency, err = query.AddOrUpdateCurrency(db,
				model.NewCurrency(
					model.CurrencyName(currencyName),
					model.CurrencyName(displayName),
					model.Int(currencyType),
					model.Int(0), model.Int(crypto),
					model.Int(precision),
				),
			)
			if err != nil {
				log.WithError(err).Error("Failed to AddOrUpdateCurrency")
				return err
			}
		}

		result = common.CurrencyInfo{
			Name:             string(currency.Name),
			DisplayName:      string(currency.DisplayName),
			Available:        currency.IsAvailable(),
			AutoCreate:       currency.AutoCreate,
			Type:             common.CurrencyType(currency.GetType()),
			Crypto:           currency.IsCrypto(),
			DisplayPrecision: uint(currency.DisplayPrecision()),
		}

		return nil
	})

	if err == nil {
		log.WithFields(logrus.Fields{
			"Name":             result.Name,
			"DisplayName":      result.DisplayName,
			"Available":        result.Available,
			"AutoCreate":       result.AutoCreate,
			"Type":             result.Type,
			"Crypto":           result.Crypto,
			"DisplayPrecision": result.DisplayPrecision,
		}).Warn("Currency created")
	}

	return result, err
}

func OnCurrencyCreate(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Currencying.OnCurrencyCreate")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CurrencyInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Name": request.Name,
			})

			currency, err := CurrencyCreate(ctx, request.Name, request.DisplayName, request.Type, request.Crypto, request.DisplayPrecision)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CurrencyCreate")
				return nil, cache.ErrInternalError
			}

			log.Info("Currency Created")

			// create & return response
			result := currency
			return &result, nil
		})
}
