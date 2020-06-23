// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"errors"
	"time"

	accounting "github.com/condensat/bank-core/accounting/client"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrProcessingBatchWithdraw = errors.New("Error Processing BatchWithdraw")
)

func BatchWithdraw(ctx context.Context, epoch time.Time, chains []string) {
	processBatchWithdraw(ctx, epoch, chains)
}

func processBatchWithdraw(ctx context.Context, epoch time.Time, chains []string) {
	log := logger.Logger(ctx).WithField("Method", "tasks.processBatchWithdraw")
	log = log.WithField("Epoch", epoch)

	for _, chain := range chains {
		log = log.WithField("Chain", chain)
		log.Debugf("Process Batch Withdraw")

		err := processBatchWithdrawChain(ctx, chain)
		if err != nil {
			log.WithError(err).Error("Failed to processBatchWithdrawChain")
			continue
		}
	}
}

func processBatchWithdrawChain(ctx context.Context, chain string) error {
	log := logger.Logger(ctx).WithField("Method", "tasks.processBatchWithdrawChain")

	// Acquire Lock
	lock, err := cache.LockBatchNetwork(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock batchNetwork")
		return ErrProcessingBatchWithdraw
	}
	defer lock.Unlock()

	list, err := accounting.BatchWithdrawList(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to get BatchWithdrawList from accounting")
		return ErrProcessingBatchWithdraw
	}

	log.WithField("Count", len(list.Batches)).
		Debugf("BatchWithdraws to process")

	for _, batch := range list.Batches {
		if batch.Network != chain {
			log.Warn("Invalid Batch Network")
			continue
		}
		if len(batch.Withdraws) == 0 {
			log.Warn("Empty Batch withdraws")
			continue
		}

		log.
			WithFields(logrus.Fields{
				"BatchID": batch.BatchID,
				"Network": batch.Network,
				"Status":  batch.Status,
				"Count":   len(batch.Withdraws),
			}).Debug("Processing Batch")

		// Todo: process batch withdraw

	}
	return nil
}
