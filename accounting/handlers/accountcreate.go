// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

func AccountCreate(ctx context.Context, userID uint64, info common.AccountInfo) (common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountCreate")
	var result common.AccountInfo

	log = log.WithField("UserID", userID)

	// Acquire Lock
	lock, err := cache.LockUser(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock user")
		return result, cache.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db database.Context) error {

		account, err := query.CreateAccount(db, model.Account{
			UserID:       model.UserID(userID),
			CurrencyName: model.CurrencyName(info.Currency.Name),
			Name:         model.AccountName(info.Name),
		})
		if err != nil {
			return err
		}

		status, err := query.AddOrUpdateAccountState(db, model.AccountState{
			AccountID: account.ID,
			State:     model.AccountStatusCreated,
		})
		if err != nil {
			return err
		}

		result = common.AccountInfo{
			AccountID: uint64(account.ID),
			Currency: common.CurrencyInfo{
				Name: string(account.CurrencyName),
			},
			Name:   string(account.Name),
			Status: string(status.State),
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

func OnAccountCreate(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountCreate")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountCreation
	return messaging.HandleRequest(ctx, appcontext.AppName(ctx), message, &request,
		func(ctx context.Context, _ messaging.BankObject) (messaging.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"UserID":   request.UserID,
				"Currency": request.Info.Currency,
				"Name":     request.Info.Name,
				"Status":   request.Info.Status,
			})

			account, err := AccountCreate(ctx, request.UserID, request.Info)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AccountCreate")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"AccountID": account.AccountID,
			})

			log.Info("Account Created")

			// create & return response
			return &common.AccountCreation{
				UserID: request.UserID,
				Info:   account,
			}, nil
		})
}
