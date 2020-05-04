// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/accounting/client"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/sirupsen/logrus"
)

// UpdateOperations
func UpdateOperations(ctx context.Context, epoch time.Time, chains []string) {
	log := logger.Logger(ctx).WithField("Method", "task.ChainUpdate")
	db := appcontext.Database(ctx)

	activeStatuses, err := database.FindActiveOperationStatus(db)
	if err != nil {
		log.WithError(err).
			Error("Failed to FindActiveOperationInfo")
		return
	}

	for _, status := range activeStatuses {
		// skip up to date statuses
		if status.State == status.Accounted {
			continue
		}

		addr, operation, err := getOperationInfos(db, status.OperationInfoID)
		if err != nil {
			log.WithError(err).
				Error("Failed to getOperationInfos")
			continue
		}

		// deposit amount to account
		accountDeposit := client.AccountDepositSync
		accountedStatus := "settled"
		switch status.State {

		case "received":
			accountDeposit = client.AccountDepositAsyncStart
			accountedStatus = "received"

		case "confirmed":
			// sync if directly confirmed (previous state empty)
			if status.Accounted == "received" {
				// End async operation
				accountDeposit = client.AccountDepositAsyncEnd
				accountedStatus = "settled"
			}
		}
		accountEntry, err := accountDeposit(ctx, uint64(addr.AccountID), uint64(operation.ID), float64(operation.Amount), "WalletDeposit")
		if err != nil {
			log.WithError(err).
				Error("Failed to AccountDeposit")
			continue
		}

		log.WithFields(logrus.Fields{
			"AccountID":        accountEntry.AccountID,
			"Accounted":        accountedStatus,
			"State":            status.State,
			"TxID":             operation.TxID,
			"Currency":         accountEntry.Currency,
			"ReferenceID":      accountEntry.ReferenceID,
			"OperationType":    accountEntry.OperationType,
			"SynchroneousType": accountEntry.SynchroneousType,
		}).Info("Wallet Deposit")

		// update Accounted status
		status.Accounted = accountedStatus
		if status.Accounted == "settled" {
			status.State = accountedStatus
		}
		_, err = database.AddOrUpdateOperationStatus(db, status)
		if err != nil {
			log.WithError(err).
				Error("Failed to AddOrUpdateOperationStatus")
			continue
		}
	}

	log.WithFields(logrus.Fields{
		"Epoch": epoch.Truncate(time.Millisecond),
	}).Info("Operations updated")
}

func getOperationInfos(db bank.Database, operationInfoID model.OperationInfoID) (model.CryptoAddress, model.OperationInfo, error) {
	// fetch OperationInfo from db
	operation, err := database.GetOperationInfo(db, operationInfoID)
	if err != nil {
		return model.CryptoAddress{}, model.OperationInfo{}, err
	}

	// fetch CryptoAddress from db
	addr, err := database.GetCryptoAddress(db, operation.CryptoAddressID)
	if err != nil {
		return model.CryptoAddress{}, model.OperationInfo{}, err
	}

	return addr, operation, nil
}
