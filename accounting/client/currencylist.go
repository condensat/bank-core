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

func CurrencyList(ctx context.Context) (common.CurrencyList, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyList")

	var result common.CurrencyList
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CurrencyListSubject, &result, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyList{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(result.Currencies),
	}).Debug("Currency Created")

	return result, nil
}
