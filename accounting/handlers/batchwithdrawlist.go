// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func BatchWithdrawList(ctx context.Context, status, network string) (common.BatchWithdraws, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.BatchWithdrawList")

	// Database Query
	db := appcontext.Database(ctx)
	batches, err := database.GetLastBatchInfoByStatusAndNetwork(db, model.BatchStatus(status), model.BatchNetwork(network))
	if err != nil {
		log.WithError(err).
			Error("Failed to GetLastBatchInfoByStatusAndNetwork")
	}

	result := common.BatchWithdraws{
		Network: network,
	}
	for _, batch := range batches {
		if batch.Type != model.BatchInfoCrypto {
			log.Warn("Wrong Batch Type")
			continue
		}

		withdraws, err := database.GetBatchWithdraws(db, batch.BatchID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetBatchWithdraws")
			continue
		}

		batchWithdraw := common.BatchWithdraw{
			BatchID: uint64(batch.BatchID),
			Network: network,
			Status:  string(batch.Status),
		}

		for _, wID := range withdraws {
			w, err := database.GetWithdraw(db, wID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetWithdraw")
				continue
			}

			if w.Amount == nil || *w.Amount == 0.0 {
				log.WithError(err).
					Error("Invalid withdraw amount")
				continue
			}
			wt, err := database.GetWithdrawTargetByWithdrawID(db, wID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetWithdrawTargetByWithdrawID")
				continue
			}

			data, err := wt.OnChainData()
			if err != nil {
				log.WithError(err).
					Error("Failed to get OnChainData")
				continue
			}
			batchWithdraw.Withdraws = append(batchWithdraw.Withdraws, common.WithdrawInfo{
				Amount:    float64(*w.Amount),
				PublicKey: data.PublicKey,
			})
		}

		if len(batchWithdraw.Withdraws) != len(withdraws) {
			log.WithFields(logrus.Fields{
				"Count":    len(batchWithdraw.Withdraws),
				"Expected": len(withdraws),
			}).Error("Somme withdraws are missing")
			// Todo: cancel batch
			continue
		}

		result.Batches = append(result.Batches, batchWithdraw)
	}

	return result, err
}

func OnBatchWithdrawList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnBatchWithdrawList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.BatchWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Network": request.Network,
				"Status":  request.Status,
			})

			response, err := BatchWithdrawList(ctx, request.Status, request.Network)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to list batch withdraws")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &response, nil
		})
}
