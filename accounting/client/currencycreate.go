// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencyCreate(ctx context.Context, currencyName, currencyDisplayName string, currencyType common.CurrencyType, isCrypto bool, displayPrecision uint) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyCreate")

	request := common.CurrencyInfo{
		Name:             currencyName,
		DisplayName:      currencyDisplayName,
		Available:        false,
		AutoCreate:       false,
		Crypto:           isCrypto,
		Type:             currencyType,
		Asset:            currencyType == 2,
		DisplayPrecision: displayPrecision,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CurrencyCreateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Name":             result.Name,
		"DisplayName":      result.DisplayName,
		"Available":        result.Available,
		"AutoCreate":       result.AutoCreate,
		"Crypto":           result.Crypto,
		"Type":             result.Type,
		"Asset":            result.Asset,
		"DisplayPrecision": result.DisplayPrecision,
	}).Debug("Currency Created")

	return result, nil
}
