// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/accounting/handlers"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
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
	nats.SubscribeWorkers(ctx, common.BatchWithdrawUpdateSubject, 2*concurencyLevel, handlers.OnBatchWithdrawUpdate)

	nats.SubscribeWorkers(ctx, common.UserWithdrawListSubject, 2*concurencyLevel, handlers.OnUserWithdrawList)

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

	// Initialize SingleCall nonce
	const singleCallName = "accounting.scheduledWithdrawBatch"
	err := cache.InitSingleCall(ctx, singleCallName)
	if err != nil {
		log.WithError(err).
			Panic("Failed to InitSingleCall")
	}

	log.Info("Start batch Scheduler")

	for epoch := range utils.Scheduler(ctx, interval, delay) {
		_ = cache.ExecuteSingleCall(ctx, singleCallName,

			// Single execution among all services
			func(ctx context.Context) error {

				log := log.WithFields(logrus.Fields{
					"Epoch": epoch.Truncate(time.Millisecond),
				})

				err := processPendingWithdraws(ctx)
				if err != nil {
					log.WithError(err).
						Error("Failed to processPendingWithdraws")
					// continue to next task
				}

				err = processPendingBatches(ctx)
				if err != nil {
					log.WithError(err).
						Error("Failed to processPendingBatches")
					// continue to next task
				}

				err = processConfirmedBatches(ctx)
				if err != nil {
					log.WithError(err).
						Error("Failed to processConfirmedBatches")
					// continue to next task
				}

				return nil
			})
	}
}

func processPendingWithdraws(ctx context.Context) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processPendingWithdraws")

	withdraws, err := FetchCreatedWithdraws(ctx)
	if err != nil {
		log.WithError(err).
			Error("Failed to FetchCreatedWithdraws")
		return err
	}

	if len(withdraws) == 0 {
		log.
			Debug("FetchCreatedWithdraws returns empty withdraw target")
		return err
	}

	err = ProcessWithdraws(ctx, withdraws)
	if err != nil {
		log.WithError(err).
			Error("Failed to ProcessWithdraws")
		return err
	}

	return nil
}

func processPendingBatches(ctx context.Context) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processPendingBatches")
	db := appcontext.Database(ctx)

	log.Info("Process batches")

	batches, err := database.FetchBatchReady(db)
	if err != nil {
		log.WithError(err).
			Error("Failed to ProcessWithdraws")
		return err
	}

	for _, batch := range batches {
		if !batch.IsComplete() {
			continue
		}
		info, err := database.GetLastBatchInfo(db, batch.ID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetLastBatchInfo")
			continue
		}
		if info.Status != model.BatchStatusCreated {
			log.
				Warning("Batch status is not BatchStatusCreated")
			continue
		}

		_, err = database.AddBatchInfo(db, batch.ID, model.BatchStatusReady, info.Type, info.Data)
		if err != nil {
			log.WithError(err).
				Error("Failed to AddBatchInfo")
			continue
		}
	}

	return nil
}

func processConfirmedBatches(ctx context.Context) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processConfirmedBatches")
	db := appcontext.Database(ctx)

	log.Info("Process batches")

	batches, err := database.FetchBatchByLastStatus(db, model.BatchStatusConfirmed)
	if err != nil {
		log.WithError(err).
			Error("Failed to FetchBatchByLastStatus")
		return err
	}

	for _, batch := range batches {
		info, err := database.GetLastBatchInfo(db, batch.ID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetLastBatchInfo")
			continue
		}
		if info.Status != model.BatchStatusConfirmed {
			log.
				Warning("Batch status is not BatchStatusConfirmed")
			continue
		}

		// within a db transaction
		err = db.Transaction(func(db bank.Database) error {
			// Mark WithdrawInfo as settled
			withdraws, err := database.GetBatchWithdraws(db, batch.ID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetBatchWithdraws")
				return err
			}
			for _, wID := range withdraws {
				_, err = database.AddWithdrawInfo(db, wID, model.WithdrawStatusSettled, "{}")
				if err != nil {
					log.WithError(err).
						Error("Failed to AddWithdrawInfo")
					return err
				}
			}

			// Mark BatchInfo as settled
			_, err = database.AddBatchInfo(db, batch.ID, model.BatchStatusSettled, info.Type, info.Data)
			if err != nil {
				log.WithError(err).
					Error("Failed to AddBatchInfo")
				return err
			}

			// Settle account operation
			for _, wID := range withdraws {
				// withdrawID is use as reference in transfert account operation
				err = settleAccountOperation(ctx, db, model.RefID(wID))
				if err != nil {
					log.WithError(err).
						Error("Failed to settleAccountOperation")
					return err
				}
			}

			return nil
		})

		if err != nil {
			log.WithError(err).
				Error("Failed to settle confirmed batch")
			continue
		}
	}

	return nil
}

func settleAccountOperation(ctx context.Context, db bank.Database, refID model.RefID) error {
	// Find transfer account operation
	op, err := database.FindAccountOperationByReference(db, model.SynchroneousTypeAsyncStart, model.OperationTypeTransfer, refID)
	if err != nil {
		return err
	}

	// Acquire Lock
	lock, err := cache.LockAccount(ctx, uint64(op.AccountID))
	if err != nil {
		return nil
	}
	defer lock.Unlock()

	// create new AccountOperation, removing and unlock amount
	_, err = database.TxAppendAccountOperation(db, model.NewAccountOperation(0,
		op.AccountID,
		model.SynchroneousTypeAsyncEnd,
		model.OperationTypeWithdraw,
		op.ReferenceID,
		time.Now(),
		-(*op.Amount), 0.0,
		-(*op.LockAmount), 0.0,
	))
	return err
}
