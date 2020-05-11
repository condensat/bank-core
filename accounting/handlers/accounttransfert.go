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

func AccountTransfert(ctx context.Context, transfert common.AccountTransfert) (common.AccountTransfert, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountTransfert")

	log = log.WithFields(logrus.Fields{
		"SrcAccountID": transfert.Source.AccountID,
		"DstAccountID": transfert.Destination.AccountID,
		"Currency":     transfert.Source.Currency,
		"Amount":       transfert.Source.Amount,
	})

	// check for accounts
	if transfert.Source.AccountID == transfert.Destination.AccountID {
		log.
			Error("Can not transfert within same account")
		return common.AccountTransfert{}, database.ErrInvalidAccountOperation
	}

	db := appcontext.Database(ctx)

	// check for currencies match
	{
		// fetch source account from DB
		srcAccount, err := database.GetAccountByID(db, model.AccountID(transfert.Source.AccountID))
		if err != nil {
			log.WithError(err).
				Error("Failed to get srcAccount")
			return common.AccountTransfert{}, database.ErrInvalidAccountOperation
		}
		// fetch destination account from DB
		dstAccount, err := database.GetAccountByID(db, model.AccountID(transfert.Destination.AccountID))
		if err != nil {
			log.WithError(err).
				Error("Failed to get dstAccount")
			return common.AccountTransfert{}, database.ErrInvalidAccountOperation
		}
		// currency must match
		if srcAccount.CurrencyName != dstAccount.CurrencyName {
			log.WithFields(logrus.Fields{
				"SrcCurrency": srcAccount.CurrencyName,
				"DstCurrency": dstAccount.CurrencyName,
			}).Error("Can not transfert currencies")
			return common.AccountTransfert{}, database.ErrInvalidAccountOperation
		}
	}

	// Acquire Locks for source and destination accounts
	lockSource, err := cache.LockAccount(ctx, transfert.Source.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountTransfert{}, cache.ErrLockError
	}
	defer lockSource.Unlock()

	lockDestination, err := cache.LockAccount(ctx, transfert.Destination.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountTransfert{}, cache.ErrLockError
	}
	defer lockDestination.Unlock()

	// Prepare data
	transfert.Source.SynchroneousType = transfert.Destination.SynchroneousType
	transfert.Source.OperationType = transfert.Destination.OperationType
	transfert.Source.ReferenceID = transfert.Destination.ReferenceID
	transfert.Source.Timestamp = transfert.Destination.Timestamp
	transfert.Source.Amount = -transfert.Destination.Amount // do not create money
	transfert.Source.Label = transfert.Destination.Label

	// Store operations
	operations, err := database.AppendAccountOperationSlice(db,
		convertEntryToOperation(transfert.Source),
		convertEntryToOperation(transfert.Destination),
	)
	if err != nil {
		log.WithError(err).
			Error("Failed to AppendAccountOperationSlice")
		return common.AccountTransfert{}, err
	}

	// response should contains 2 operations
	if len(operations) != 2 {
		log.
			Error("Invalid operations count")
		return common.AccountTransfert{}, database.ErrInvalidAccountOperation
	}

	source := operations[0]
	destination := operations[1]
	log.WithFields(logrus.Fields{
		"SrcPrevID": source.PrevID,
		"DstPrevID": destination.PrevID,
	}).Trace("Account transfert")

	return common.AccountTransfert{
		Source:      convertOperationToEntry(source, "N/A"),
		Destination: convertOperationToEntry(destination, "N/A"),
	}, nil
}

func OnAccountTransfert(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountTransfert")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountTransfert
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {

			response, err := AccountTransfert(ctx, request)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"SrcAccountID": request.Source.AccountID,
						"DstAccountID": request.Destination.AccountID,
					}).Errorf("Failed to AccountTransfert")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}

// conversion helpers

func convertOperationToEntry(op model.AccountOperation, label string) common.AccountEntry {
	return common.AccountEntry{

		OperationID:     uint64(op.ID),
		OperationPrevID: uint64(op.PrevID),

		AccountID:        uint64(op.AccountID),
		ReferenceID:      uint64(op.ReferenceID),
		OperationType:    string(op.OperationType),
		SynchroneousType: string(op.SynchroneousType),

		Timestamp: op.Timestamp,
		Label:     label,
		Amount:    float64(*op.Amount),
		Balance:   float64(*op.Balance),

		LockAmount:  float64(*op.LockAmount),
		TotalLocked: float64(*op.TotalLocked),
	}
}

func convertEntryToOperation(entry common.AccountEntry) model.AccountOperation {
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

	return model.AccountOperation{
		AccountID:        model.AccountID(entry.AccountID),
		SynchroneousType: model.ParseSynchroneousType(entry.SynchroneousType),
		OperationType:    model.ParseOperationType(entry.OperationType),
		ReferenceID:      model.RefID(entry.ReferenceID),

		Amount:  &amount,
		Balance: &balance,

		LockAmount:  &lockAmount,
		TotalLocked: &totalLocked,

		Timestamp: entry.Timestamp,
	}
}
