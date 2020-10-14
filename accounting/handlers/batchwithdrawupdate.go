// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/database/encoding"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

func BatchWithdrawUpdate(ctx context.Context, batchID uint64, status, txID string, height int) (common.BatchStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.BatchWithdrawUpdate")

	// Database Query
	db := appcontext.Database(ctx)

	batchInfo, err := query.GetLastBatchInfo(db, model.BatchID(batchID))
	if err != nil {
		log.WithError(err).
			Error("Failed to GetLastBatchInfo")
		return common.BatchStatus{}, err
	}
	if !canUpdateStatus(batchInfo.Status, model.BatchStatus(status)) {
		log.WithError(err).
			WithField("From", batchInfo.Status).
			WithField("To", status).
			Error("Can not update Status")
		return common.BatchStatus{}, err
	}
	if height < 0 {
		log.Error("Invalid Height")
		return common.BatchStatus{}, errors.New("Invalid Height")
	}

	// change status to processing, with TxID
	data, err := encoding.EncodeData(&model.BatchInfoCryptoData{
		TxID:   model.String(txID),
		Height: model.Int(height),
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to EncodeData")
		return common.BatchStatus{}, err
	}

	batchInfo, err = query.AddBatchInfo(db, batchInfo.BatchID, model.BatchStatus(status), model.BatchInfoCrypto, model.BatchInfoData(data))
	if err != nil {
		log.WithError(err).
			Error("Failed to AddBatchInfos")
		return common.BatchStatus{}, err
	}

	return common.BatchStatus{
		BatchID: batchID,
		Status:  string(batchInfo.Status),
	}, err
}

func OnBatchWithdrawUpdate(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnBatchWithdrawUpdate")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.BatchUpdate
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"BatchID": request.BatchID,
				"Status":  request.Status,
				"TxID":    request.TxID,
				"Height":  request.Height,
			})

			response, err := BatchWithdrawUpdate(ctx, request.BatchID, request.Status, request.TxID, request.Height)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to update batch withdraws")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &response, nil
		})
}

func canUpdateStatus(from, to model.BatchStatus) bool {
	if from == to {
		return false
	}

	switch to {
	case model.BatchStatusReady:
		return from == model.BatchStatusCreated
	case model.BatchStatusProcessing:
		return from == model.BatchStatusReady
	case model.BatchStatusConfirmed:
		return from == model.BatchStatusProcessing
	case model.BatchStatusSettled:
		return from == model.BatchStatusConfirmed
	case model.BatchStatusCanceled:
		return from != model.BatchStatusProcessing

	default:
		return false
	}
}
