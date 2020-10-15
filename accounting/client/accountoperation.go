// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func AccountDepositSync(ctx context.Context, accountID, referenceID uint64, amount float64, label string) (common.AccountEntry, error) {
	return accountDeposit(ctx, "sync", accountID, referenceID, amount, label)
}

func AccountDepositAsyncStart(ctx context.Context, accountID, referenceID uint64, amount float64, label string) (common.AccountEntry, error) {
	return accountDeposit(ctx, "async-start", accountID, referenceID, amount, label)
}

func AccountDepositAsyncEnd(ctx context.Context, accountID, referenceID uint64, amount float64, label string) (common.AccountEntry, error) {
	return accountDeposit(ctx, "async-end", accountID, referenceID, amount, label)
}

func accountDeposit(ctx context.Context, sync string, accountID, referenceID uint64, amount float64, label string) (common.AccountEntry, error) {
	if len(sync) == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}
	if accountID == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	// Deposit amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	var lockAmount float64
	switch sync {
	case "sync":
		// amount ready
	case "async-start":
		// lock amount
		lockAmount = amount
		// amount ready
	case "async-end":
		// unlock amount
		lockAmount = -amount
		amount = 0.0
	}

	return accountDepositRequest(ctx, common.AccountEntry{
		AccountID: accountID,

		ReferenceID:      referenceID,
		OperationType:    "deposit",
		SynchroneousType: sync,
		Timestamp:        time.Now(),

		Label: label,

		Amount:     amount,
		LockAmount: lockAmount,
	})
}

func accountDepositRequest(ctx context.Context, entry common.AccountEntry) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountDeposit")
	log = log.WithField("AccountID", entry.AccountID)

	var result common.AccountEntry
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.AccountOperationSubject, &entry, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountEntry{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"OperationID":     result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
	}).Debug("Account amount")

	return result, nil
}

func AccountWithdraw(ctx context.Context, accountID, referenceID uint64, amount float64, label string) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountWithdraw")

	if accountID == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	// Deposit amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	log = log.WithField("AccountID", accountID)

	request := common.AccountEntry{
		AccountID: accountID,

		ReferenceID:      referenceID,
		OperationType:    "withdraw",
		SynchroneousType: "sync",
		Timestamp:        time.Now(),

		Label: label,

		Amount:     -amount, // withdraw remove amount from account
		LockAmount: 0.0,     // no lock on withdraw
	}

	var result common.AccountEntry
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.AccountOperationSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountEntry{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"OperationID":     result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
	}).Debug("Account Withdraw")

	return result, nil
}

func AccountTransfer(ctx context.Context, srcAccountID, dstAccountID, referenceID uint64, currency string, amount float64, label string) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountTransfer")

	if srcAccountID == 0 || dstAccountID == 0 {
		return common.AccountTransfer{}, cache.ErrInternalError
	}
	if srcAccountID == dstAccountID {
		return common.AccountTransfer{}, cache.ErrInternalError
	}

	// currency must be valid
	if len(currency) == 0 {
		return common.AccountTransfer{}, cache.ErrInternalError
	}

	// deposit amount must be positive
	if amount <= 0.0 {
		return common.AccountTransfer{}, cache.ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"SrcAccountID": srcAccountID,
		"DstAccountID": dstAccountID,

		"Amount":   amount,
		"Currency": currency,
	})

	request := common.AccountTransfer{
		Source: common.AccountEntry{
			AccountID: srcAccountID,
			Currency:  currency,
		},
		Destination: common.AccountEntry{
			AccountID: dstAccountID,

			OperationType:    "transfer",
			SynchroneousType: "sync",
			ReferenceID:      referenceID,

			Timestamp: time.Now(),
			Amount:    amount,

			Label: label,

			LockAmount: 0.0, // no lock on sync account transfer
			Currency:   currency,
		},
	}

	var result common.AccountTransfer
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.AccountTransferSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountTransfer{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SrcID":      result.Source.OperationID,
		"SrcPrevID":  result.Source.OperationPrevID,
		"SrcBalance": result.Source.Balance,

		"DstID":      result.Destination.OperationID,
		"DstPrevID":  result.Destination.OperationPrevID,
		"DstBalance": result.Destination.Balance,
	}).Debug("Account amount")

	return result, nil
}
