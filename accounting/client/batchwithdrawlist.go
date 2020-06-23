// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
)

func BatchWithdrawList(ctx context.Context, network string) (common.BatchWithdraws, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.BatchWithdrawList")
	log = log.WithField("Network", network)

	request := common.BatchWithdraw{
		Network: network,
	}

	var result common.BatchWithdraws
	err := messaging.RequestMessage(ctx, common.BatchWithdrawListSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.BatchWithdraws{}, messaging.ErrRequestFailed
	}

	log.WithField("Count", len(result.Batches)).
		Debug("BatchWithdraw List")

	return result, nil
}
