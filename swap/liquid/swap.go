// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package liquid

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/condensat/bank-core/swap/liquid/common"
	"github.com/condensat/bank-core/swap/liquid/handlers"

	"github.com/sirupsen/logrus"
)

type Swap int

func (p *Swap) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Swap.Run")

	p.registerHandlers(cache.RedisMutexContext(ctx))

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Liquid Swap Service started")

	<-ctx.Done()
}

func (p *Swap) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 8

	nats.SubscribeWorkers(ctx, common.SwapCreateProposalSubject, 2*concurencyLevel, handlers.OnCreateSwapProposal)
	nats.SubscribeWorkers(ctx, common.SwapInfoProposalSubject, 2*concurencyLevel, handlers.OnInfoSwapProposal)
	nats.SubscribeWorkers(ctx, common.SwapFinalizeProposalSubject, 2*concurencyLevel, handlers.OnFinalizeSwapProposal)
	nats.SubscribeWorkers(ctx, common.SwapAcceptProposalSubject, 2*concurencyLevel, handlers.OnAcceptSwapProposal)

	log.Debug("Liquid Swap registered")
}
