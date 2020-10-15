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

func CurrencySetAvailable(ctx context.Context, currencyName string, available bool) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencySetAvailable")

	request := common.CurrencyInfo{
		Name:      currencyName,
		Available: available,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CurrencySetAvailableSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Name":      result.Name,
		"Available": result.Available,
	}).Debug("Currency SetAvailable")

	return result, nil
}
