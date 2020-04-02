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

func AccountCreate(ctx context.Context, userID uint64, currency string) (common.AccountCreation, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountCreate")

	request := common.AccountCreation{
		UserID: userID,
		Info: common.AccountInfo{
			Currency: currency,
		},
	}

	var result common.AccountCreation
	err := messaging.RequestMessage(ctx, common.AccountCreateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountCreation{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"UserID":    result.UserID,
		"AccountID": result.Info.AccountID,
	}).Debug("Account Created")

	return result, nil
}
