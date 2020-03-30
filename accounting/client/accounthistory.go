// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"
	"time"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func AccountHistory(ctx context.Context, accountID uint64, from, to time.Time) (common.AccountHistory, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountHistory")

	request := common.AccountHistory{
		AccountID: accountID,
		From:      from,
		To:        to,
	}

	var result common.AccountHistory
	err := messaging.RequestMessage(ctx, common.AccountHistorySubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountHistory{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"AccountID": result.AccountID,
		"Count":     len(result.History),
	}).Debug("Account History")

	return result, nil
}
