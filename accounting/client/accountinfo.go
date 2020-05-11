// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func AccountInfo(ctx context.Context, accountID uint64) (common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountInfo")

	request := common.AccountInfo{
		AccountID: accountID,
	}

	var result common.AccountInfo
	err := messaging.RequestMessage(ctx, common.AccountInfoSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"AccountID": result.AccountID,
		"Currency":  result.Currency,
		"Name":      result.Name,
		"Status":    result.Status,
	}).Debug("Account Info")

	return result, nil
}
