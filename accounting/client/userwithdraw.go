// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func UserWithdrawsCrypto(ctx context.Context, userID uint64) (common.UserWithdraws, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.UserWithdrawsCrypto")
	log = log.WithField("UserID", userID)

	if userID == 0 {
		return common.UserWithdraws{}, cache.ErrInternalError
	}

	var result common.UserWithdraws
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.UserWithdrawListSubject, &common.UserWithdraws{UserID: userID}, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.UserWithdraws{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(result.Withdraws),
	}).Debug("UserWithdraws request")

	return result, nil
}
