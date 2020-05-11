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

func CurrencyInfo(ctx context.Context, currencyName string) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CurrencyInfo")
	var result common.CurrencyInfo

	log = log.WithField("CurrencyName", currencyName)

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		// check if currency exists
		currency, err := database.GetCurrencyByName(db, model.CurrencyName(currencyName))
		if err != nil {
			log.WithError(err).Error("Failed to GetCurrencyByName")
			return err
		}

		if string(currency.Name) != currencyName {
			return database.ErrCurrencyNotFound
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
			"Name":        result.Name,
			"DisplayName": result.DisplayName,
			"Available":   result.Available,
			"AutoCreate":  result.AutoCreate,
			"Type":        result.Type,
			"Crypto":      result.Crypto,
		}).Debug("Currency info")
	}

	return result, err
}

func OnCurrencyInfo(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Currencying.OnCurrencyInfo")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CurrencyInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Name": request.Name,
			})

			currency, err := CurrencyInfo(ctx, request.Name)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CurrencyInfo")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.CurrencyInfo{
				Name:             currency.Name,
				DisplayName:      currency.DisplayName,
				Available:        currency.Available,
				AutoCreate:       currency.AutoCreate,
				Type:             currency.Type,
				Crypto:           currency.Crypto,
				DisplayPrecision: currency.DisplayPrecision,
			}, nil
		})
}
