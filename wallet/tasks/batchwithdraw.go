// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"errors"
	"time"

	accounting "github.com/condensat/bank-core/accounting/client"
	"github.com/condensat/bank-core/wallet/chain"
	"github.com/condensat/bank-core/wallet/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrProcessingBatchWithdraw = errors.New("Error Processing BatchWithdraw")
)

type ChainHeightMap map[string]uint64

func BatchWithdraw(ctx context.Context, epoch time.Time, chains []string) {
	log := logger.Logger(ctx).WithField("Method", "tasks.processBatchWithdraw")
	log = log.WithField("Epoch", epoch)

	chainsStates, err := chain.FetchChainsState(ctx, chains...)
	if err != nil {
		log.WithError(err).
			Error("Failed to FetchChainsState")
		return
	}

	chainHeight := make(ChainHeightMap)
	for _, state := range chainsStates {
		chainHeight[state.Chain] = state.Height
	}

	processBatchWithdraw(ctx, epoch, chains)
	watchBatchWithdraw(ctx, epoch, chains, chainHeight)
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

func processBatchWithdrawChain(ctx context.Context, network string) error {
	log := logger.Logger(ctx).WithField("Method", "tasks.processBatchWithdrawChain")

	// Acquire Lock
	lock, err := cache.LockBatchNetwork(ctx, network)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock batchNetwork")
		return ErrProcessingBatchWithdraw
	}
	defer lock.Unlock()

	list, err := accounting.ListBatchWithdrawReady(ctx, network)
	if err != nil {
		log.WithError(err).
			Error("Failed to get ListBatchWithdrawReady from accounting")
		return ErrProcessingBatchWithdraw
	}

	log.WithField("Count", len(list.Batches)).
		Debugf("BatchWithdraws to process")

	for _, batch := range list.Batches {
		if batch.Network != network {
			log.Warn("Invalid Batch Network")
			continue
		}
		if len(batch.Withdraws) == 0 {
			log.Warn("Empty Batch withdraws")
			continue
		}
		if batch.Status != "ready" {
			if len(batch.Withdraws) == 0 {
				log.Warn("Batch status is not ready")
				continue
			}

		}

		log.
			WithFields(logrus.Fields{
				"BatchID": batch.BatchID,
				"Network": batch.Network,
				"Status":  batch.Status,
				"Count":   len(batch.Withdraws),
			}).Debug("Processing Batch")

		// Resquest chain
		var spendInfo []common.SpendInfo
		for _, withdraw := range batch.Withdraws {
			spendInfo = append(spendInfo, common.SpendInfo{
				PublicAddress: withdraw.PublicKey,
				Amount:        withdraw.Amount,
			})
		}
		spendTx, err := chain.SpendFunds(ctx, network, spendInfo)
		if err != nil {
			log.WithError(err).
				Error("Failed to SpendFunds")
			continue
		}

		// Update batch status with TxID
		batchStatus, err := accounting.BatchWithdrawUpdate(ctx, uint64(batch.BatchID), "processing", string(spendTx.TxID))
		if err != nil {
			log.WithError(err).
				Error("Failed to BatchWithdrawUpdate")
			continue
		}

		log.WithFields(logrus.Fields{
			"Network": network,
			"BatchID": batch.BatchID,
			"Status":  batchStatus.Status,
		}).Info("Batch updated")
	}
	return nil
}

func watchBatchWithdraw(ctx context.Context, epoch time.Time, chains []string, chainHeight ChainHeightMap) {
	log := logger.Logger(ctx).WithField("Method", "tasks.watchBatchWithdraw")
	log = log.WithField("Epoch", epoch)

	for _, chain := range chains {
		log = log.WithField("Chain", chain)
		log.Debugf("Watch Batch Withdraw")

		err := watchBatchWithdrawChain(ctx, chain, chainHeight[chain])
		if err != nil {
			log.WithError(err).Error("Failed to watchBatchWithdrawChain")
			continue
		}
	}
}

func watchBatchWithdrawChain(ctx context.Context, network string, currentHeight uint64) error {
	log := logger.Logger(ctx).WithField("Method", "tasks.watchBatchWithdrawChain")

	// Acquire Lock
	lock, err := cache.LockBatchNetwork(ctx, network)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock batchNetwork")
		return ErrProcessingBatchWithdraw
	}
	defer lock.Unlock()

	list, err := accounting.ListBatchWithdrawProcessing(ctx, network)
	if err != nil {
		log.WithError(err).
			Error("Failed to get ListBatchWithdrawProcessing from accounting")
		return ErrProcessingBatchWithdraw
	}

	log.WithField("Count", len(list.Batches)).
		Debugf("BatchWithdraws to watch")

	for _, batch := range list.Batches {
		if batch.Network != network {
			log.Warn("Invalid Batch Network")
			continue
		}
		if batch.Status != "processing" {
			log.Warn("Batch status is not processing")
			continue
		}
		if len(batch.TxID) == 0 {
			log.Warn("Invalid batch txId")
			continue
		}

		log.
			WithFields(logrus.Fields{
				"BatchID": batch.BatchID,
				"Network": batch.Network,
				"Status":  batch.Status,
				"TxID":    batch.TxID,
			}).Debug("Watch Batch transaction")

		// Resquest chain
		txInfo, err := chain.GetTransaction(ctx, network, batch.TxID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetTransaction")
			continue
		}

		if txInfo.Confirmations < ConfirmedBlockCount {
			log.WithField("Confirmations", txInfo.Confirmations).
				Debug("Transaction not yet confirmed")
			continue
		}

		// Update batch status with TxID and height
		height := currentHeight - uint64(txInfo.Confirmations) + 1
		batchStatus, err := accounting.BatchWithdrawUpdateWithHeight(ctx, uint64(batch.BatchID), "confirmed", string(txInfo.TxID), int(height))
		if err != nil {
			log.WithError(err).
				Error("Failed to BatchWithdrawUpdateWithHeight")
			continue
		}

		log.WithFields(logrus.Fields{
			"Network": network,
			"BatchID": batch.BatchID,
			"Status":  batchStatus.Status,
		}).Info("Batch updated")
	}
	return nil
}
