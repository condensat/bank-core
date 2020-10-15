// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/swap/liquid/common"

	"github.com/sirupsen/logrus"
)

func AcceptSwapProposal(ctx context.Context, swapID uint64, address common.ConfidentialAddress, payload common.Payload, feeRate float64) (common.SwapProposal, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.client.AcceptSwapProposal")

	if !payload.Valid() {
		return common.SwapProposal{}, common.ErrInvalidPayload
	}

	request := common.SwapProposal{
		SwapID:  swapID,
		Address: address,
		Payload: payload,
		FeeRate: feeRate,
	}

	var result common.SwapProposal
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.SwapAcceptProposalSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.SwapProposal{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SwapID": result.SwapID,
	}).Debug("Accept SwapProposal")

	return result, nil
}
