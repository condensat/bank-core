// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/accounting/handlers"
	"github.com/condensat/bank-core/accounting/internal"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

type Accounting int

func (p *Accounting) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.Run")

	p.registerHandlers(internal.RedisMutexContext(ctx))

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Accounting Service started")

	<-ctx.Done()
}

func (p *Accounting) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	nats.SubscribeWorkers(ctx, common.CurrencyCreateSubject, 8, handlers.OnCurrencyCreate)
	nats.SubscribeWorkers(ctx, common.CurrencyListSubject, 8, handlers.OnCurrencyList)
	nats.SubscribeWorkers(ctx, common.CurrencySetAvailableSubject, 8, handlers.OnCurrencySetAvailable)

	nats.SubscribeWorkers(ctx, common.AccountCreateSubject, 8, handlers.OnAccountCreate)
	nats.SubscribeWorkers(ctx, common.AccountListSubject, 8, handlers.OnAccountList)
	nats.SubscribeWorkers(ctx, common.AccountHistorySubject, 8, handlers.OnAccountHistory)
	nats.SubscribeWorkers(ctx, common.AccountSetStatusSubject, 8, handlers.OnAccountSetStatus)
	nats.SubscribeWorkers(ctx, common.AccountOperationSubject, 16, handlers.OnAccountOperation)
	nats.SubscribeWorkers(ctx, common.AccountTransfertSubject, 16, handlers.OnAccountTransfert)

	log.Debug("Bank Accounting registered")
}
