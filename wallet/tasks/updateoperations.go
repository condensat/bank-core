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
	log := logger.Logger(ctx).WithField("Method", "tasks.UpdateOperations")
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

		userID, addr, operation, err := getOperationInfos(db, status.OperationInfoID)
		if err != nil {
			log.WithError(err).
				Error("Failed to getOperationInfos")
			continue
		}

		accountID := uint64(addr.AccountID)
		if operation.AssetID != 0 {
			// create user asset account if needed
			newAccountID, err := createUserAssetAccount(ctx, uint64(userID), uint64(addr.AccountID), operation.AssetID)
			if err != nil {
				log.WithError(err).
					Error("Failed to createUserAssetAccount")
				continue
			}
			accountID = newAccountID
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
		accountEntry, err := accountDeposit(ctx, accountID, uint64(operation.ID), float64(operation.Amount), "WalletDeposit")
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

func getOperationInfos(db bank.Database, operationInfoID model.OperationInfoID) (model.UserID, model.CryptoAddress, model.OperationInfo, error) {
	// fetch OperationInfo from db
	operation, err := database.GetOperationInfo(db, operationInfoID)
	if err != nil {
		return 0, model.CryptoAddress{}, model.OperationInfo{}, err
	}

	// fetch CryptoAddress from db
	addr, err := database.GetCryptoAddress(db, operation.CryptoAddressID)
	if err != nil {
		return 0, model.CryptoAddress{}, model.OperationInfo{}, err
	}

	account, err := database.GetAccountByID(db, addr.AccountID)
	if err != nil {
		return 0, model.CryptoAddress{}, model.OperationInfo{}, err
	}

	return account.UserID, addr, operation, nil
}

func createUserAssetAccount(ctx context.Context, userID, accountID uint64, assetID model.AssetID) (uint64, error) {
	log := logger.Logger(ctx).WithField("Method", "tasks.createUserAssetAccount")
	db := appcontext.Database(ctx)

	log = log.WithFields(logrus.Fields{
		"UserID":    userID,
		"AccountID": accountID,
		"AssetID":   assetID,
	})

	// no asset, no error
	if assetID == 0 {
		return accountID, nil
	}
	if userID == 0 {
		return 0, database.ErrInvalidUserID
	}

	// check if asset exists
	asset, err := database.GetAsset(db, assetID)
	if err != nil {
		log.WithError(err).
			Error("Asset NotFound")
		return accountID, nil
	}

	// check if account exists
	accounts, err := client.AccountList(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("AccountList failed")
		return accountID, nil
	}

	// find account with currency
	for _, account := range accounts.Accounts {
		if account.Currency.Name == string(asset.CurrencyName) {
			log.
				WithField("CurrencyName", asset.CurrencyName).
				WithField("AccountID", account.AccountID).
				Debug("Account Exists")
			return account.AccountID, nil
		}
	}

	// if currency does not exist try to create
	creation, err := client.AccountCreate(ctx, userID, string(asset.CurrencyName))
	if err != nil {
		log.WithError(err).
			Error("AccountCreate failed")
		return 0, err
	}
	account := creation.Info
	// curency is created and already available
	if account.Status == "normal" {
		return account.AccountID, nil
	}

	// activate currency
	account, err = client.AccountSetStatus(ctx, account.AccountID, "normal")
	if err != nil {
		log.WithError(err).
			Error("AccountSetStatus failed")
		return 0, err
	}

	log.
		WithField("NewAccountID", account.AccountID).
		Debug("Account Created")

	return account.AccountID, nil
}
