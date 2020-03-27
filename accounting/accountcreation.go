// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/sirupsen/logrus"
)

func CreateUserAccount(ctx context.Context, userID uint64, info AccountInfo) (AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CreateUserAccount")
	var result AccountInfo

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

		account, err := database.CreateAccount(db, model.Account{
			UserID:       model.UserID(userID),
			CurrencyName: model.CurrencyName(info.Currency),
			Name:         model.AccountName(info.Name),
		})
		if err != nil {
			return err
		}

		status, err := database.AddOrUpdateAccountState(db, model.AccountState{
			AccountID: account.ID,
			State:     model.AccountStatusCreated,
		})
		if err != nil {
			return err
		}

		result = AccountInfo{
			AccountID: uint64(account.ID),
			Currency:  string(account.CurrencyName),
			Name:      string(account.Name),
			Status:    string(status.State),
		}

		return nil
	})

	if err == nil {
		log.WithFields(logrus.Fields{
			"AccountID": result.AccountID,
			"Status":    result.Status,
		}).Debug("Account created")
	}

	return result, err
}
