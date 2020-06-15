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

func AccountOperation(ctx context.Context, entry common.AccountEntry) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountOperation")

	log = log.WithFields(logrus.Fields{
		"AccountID":        entry.AccountID,
		"Currency":         entry.Currency,
		"SynchroneousType": entry.SynchroneousType,
		"OperationType":    entry.OperationType,
		"ReferenceID":      entry.ReferenceID,
	})

	// Acquire Lock
	lock, err := cache.LockAccount(ctx, entry.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountEntry{}, cache.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	accountID := model.AccountID(entry.AccountID)
	amount := model.Float(entry.Amount)
	lockAmount := model.Float(entry.LockAmount)

	// Balance & totalLocked ar computed by database later, must be valid for pre-check
	var balance model.Float
	if balance < amount {
		balance = amount
	}
	var totalLocked model.Float
	if totalLocked < lockAmount {
		totalLocked = lockAmount
	}

	op, err := database.AppendAccountOperation(db, model.AccountOperation{
		AccountID:        accountID,
		SynchroneousType: model.ParseSynchroneousType(entry.SynchroneousType),
		OperationType:    model.ParseOperationType(entry.OperationType),
		ReferenceID:      model.RefID(entry.ReferenceID),

		Amount:  &amount,
		Balance: &balance,

		LockAmount:  &lockAmount,
		TotalLocked: &totalLocked,

		Timestamp: entry.Timestamp,
	})
	if err != nil {
		return common.AccountEntry{}, err
	}

	log.
		WithField("OperationID", op.ID).
		Trace("Account operation")

	return common.AccountEntry{
		OperationID: uint64(op.ID),

		AccountID:        uint64(op.AccountID),
		ReferenceID:      uint64(op.ReferenceID),
		OperationType:    string(op.OperationType),
		SynchroneousType: string(op.SynchroneousType),

		Timestamp: op.Timestamp,
		Label:     "N/A",
		Amount:    float64(*op.Amount),
		Balance:   float64(*op.Balance),

		LockAmount:  float64(*op.LockAmount),
		TotalLocked: float64(*op.TotalLocked),
	}, nil
}

func OnAccountOperation(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountOperation")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountEntry
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
			})

			response, err := AccountOperation(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AccountOperation")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &response, nil
		})
}
