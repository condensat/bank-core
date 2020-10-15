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

func CreateSwapProposal(ctx context.Context, swapID uint64, address common.ConfidentialAddress, proposal common.ProposalInfo, feeRate float64) (common.SwapProposal, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.client.CreateSwapProposal")

	if len(address) == 0 {
		return common.SwapProposal{}, common.ErrInvalidProposal
	}
	if !proposal.Valid() {
		return common.SwapProposal{}, common.ErrInvalidProposal
	}

	request := common.SwapProposal{
		SwapID:   swapID,
		Address:  address,
		Proposal: proposal,
		FeeRate:  feeRate,
	}

	var result common.SwapProposal
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.SwapCreateProposalSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.SwapProposal{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SwapID": result.SwapID,
	}).Debug("Create SwapProposal")

	return result, nil
}
