// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"errors"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrLockError = errors.New("Failed to acquire lock")
)

func ListUserAccounts(ctx context.Context, userID uint64) ([]AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.ListUserAccounts")
	var result []AccountInfo

	log = log.WithField("UserID", userID)

	// Acquire Lock
	lock, err := LockUser(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock user")
		return result, ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db bank.Database) error {
		accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, model.UserID(userID), "", "*")
		if err != nil {
			return err
		}

		for _, account := range accounts {
			accountState, err := database.GetAccountStatusByAccountID(db, account.ID)
			if err != nil {
				return err
			}

			result = append(result, AccountInfo{
				AccountID: uint64(account.ID),
				Currency:  string(account.CurrencyName),
				Name:      string(account.Name),
				Status:    string(accountState.State),
			})
		}

		return nil
	})

	if err == nil {
		log.WithField("Count", len(result)).
			Debug("User accounts retrieved")
	}

	return result, err
}

func GetAccountHistory(ctx context.Context, accountID uint64, from, to time.Time) ([]AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.GetAccountHistory")
	var result []AccountEntry

	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
		"From":      from,
		"To":        to,
	})

	// Acquire Lock
	lock, err := LockAccount(ctx, accountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return result, ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db bank.Database) error {
		account, err := database.GetAccountByID(db, model.AccountID(accountID))
		if err != nil {
			return err
		}

		operations, err := database.GeAccountHistoryRange(db, account.ID, from, to)
		if err != nil {
			return err
		}

		for _, op := range operations {
			if !op.IsValid() {
				return database.ErrInvalidAccountOperation
			}

			result = append(result, AccountEntry{
				AccountID: uint64(op.AccountID),
				Currency:  string(account.CurrencyName),

				OperationType:    string(op.OperationType),
				SynchroneousType: string(op.SynchroneousType),

				Timestamp: op.Timestamp,
				Label:     "N/A",
				Amount:    float64(*op.Amount),
				Balance:   float64(*op.Balance),

				LockAmount:  float64(*op.LockAmount),
				TotalLocked: float64(*op.TotalLocked),
			})
		}

		return nil
	})

	if err == nil {
		log.
			WithField("Count", len(result)).
			Debug("Account history retrieved")
	}

	return result, err
}
