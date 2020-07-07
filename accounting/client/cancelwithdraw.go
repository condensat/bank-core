// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
)

func CancelWithdraw(ctx context.Context, withdrawID uint64) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CancelWithdraw")
	log = log.WithField("UserID", withdrawID)

	if withdrawID == 0 {
		return common.WithdrawInfo{}, cache.ErrInternalError
	}

	var result common.WithdrawInfo
	err := messaging.RequestMessage(ctx, common.CancelWithdrawSubject, &common.WithdrawInfo{WithdrawID: withdrawID}, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.WithdrawInfo{}, messaging.ErrRequestFailed
	}

	return result, nil
}
