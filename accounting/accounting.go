// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/accounting/handlers"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

type Accounting int

const (
	DefaultInterval time.Duration = 30 * time.Second
	DefaultDelay    time.Duration = 0 * time.Second
)

func (p *Accounting) Run(ctx context.Context, bankUser model.User) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.Run")
	ctx = common.BankUserContext(ctx, bankUser)
	ctx = cache.RedisMutexContext(ctx)

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Accounting Service started")

	go p.scheduledWithdrawBatch(ctx, DefaultInterval, DefaultDelay)

	<-ctx.Done()
}

func (p *Accounting) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 8

	nats.SubscribeWorkers(ctx, common.CurrencyInfoSubject, 2*concurencyLevel, handlers.OnCurrencyInfo)
	nats.SubscribeWorkers(ctx, common.CurrencyCreateSubject, 2*concurencyLevel, handlers.OnCurrencyCreate)
	nats.SubscribeWorkers(ctx, common.CurrencyListSubject, 2*concurencyLevel, handlers.OnCurrencyList)
	nats.SubscribeWorkers(ctx, common.CurrencySetAvailableSubject, 2*concurencyLevel, handlers.OnCurrencySetAvailable)

	nats.SubscribeWorkers(ctx, common.AccountCreateSubject, 2*concurencyLevel, handlers.OnAccountCreate)
	nats.SubscribeWorkers(ctx, common.AccountInfoSubject, 2*concurencyLevel, handlers.OnAccountInfo)
	nats.SubscribeWorkers(ctx, common.AccountListSubject, 2*concurencyLevel, handlers.OnAccountList)
	nats.SubscribeWorkers(ctx, common.AccountHistorySubject, 2*concurencyLevel, handlers.OnAccountHistory)
	nats.SubscribeWorkers(ctx, common.AccountSetStatusSubject, 2*concurencyLevel, handlers.OnAccountSetStatus)
	nats.SubscribeWorkers(ctx, common.AccountOperationSubject, 8*concurencyLevel, handlers.OnAccountOperation)
	nats.SubscribeWorkers(ctx, common.AccountTransferSubject, 8*concurencyLevel, handlers.OnAccountTransfer)

	nats.SubscribeWorkers(ctx, common.AccountTransferWithdrawSubject, 2*concurencyLevel, handlers.OnAccountTransferWithdraw)

	nats.SubscribeWorkers(ctx, common.BatchWithdrawListSubject, 2*concurencyLevel, handlers.OnBatchWithdrawList)

	log.Debug("Bank Accounting registered")
}

func checkParams(interval time.Duration, delay time.Duration) (time.Duration, time.Duration) {
	if interval < time.Second {
		interval = DefaultInterval
	}
	if delay < 0 {
		delay = DefaultDelay
	}

	return interval, delay
}

func (p *Accounting) scheduledWithdrawBatch(ctx context.Context, interval time.Duration, delay time.Duration) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.scheduledWithdrawBatch")

	interval, delay = checkParams(interval, delay)

	log = log.WithFields(logrus.Fields{
		"Interval": fmt.Sprintf("%v", interval),
		"Delay":    fmt.Sprintf("%v", delay),
	})

	log.Info("Start batch Scheduler")

	for epoch := range utils.Scheduler(ctx, interval, delay) {
		log := log.WithFields(logrus.Fields{
			"Epoch": epoch.Truncate(time.Millisecond),
		})

		withdraws, err := FetchCreatedWithdraws(ctx)
		if err != nil {
			log.WithError(err).
				Error("Failed to FetchCreatedWithdraws")
			continue
		}

		if len(withdraws) == 0 {
			log.
				Debug("FetchCreatedWithdraws returns empty withdraw target")
			continue
		}

		err = ProcessWithdraws(ctx, withdraws)
		if err != nil {
			log.WithError(err).
				Error("Failed to ProcessWithdraws")
			continue
		}

		log.Debug("Withdraw processed")

		log.WithFields(logrus.Fields{
			"Epoch": epoch.Truncate(time.Millisecond),
		}).Debug("ProcessWithdraws processed")
	}
}
