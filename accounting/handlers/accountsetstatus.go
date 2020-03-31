// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/accounting/internal"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func AccountSetStatus(ctx context.Context, accountID uint64, state string) (common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountSetStatus")
	var result common.AccountInfo

	log = log.WithField("AccountID", accountID)

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		account, err := database.GetAccountByID(db, model.AccountID(accountID))
		if err != nil {
			log.WithError(err).Error("Failed to GetAccountByID")
			return err
		}

		status, err := database.GetAccountStatusByAccountID(db, model.AccountID(accountID))
		if err != nil {
			log.WithError(err).Error("Failed to GetAccountStatusByAccountID")
			return err
		}

		if string(status.State) == state {
			// NOOP
			result = common.AccountInfo{
				AccountID: uint64(account.ID),
				Status:    string(status.State),
			}
			return nil
		}

		// update acount status
		status, err = database.AddOrUpdateAccountState(db, model.AccountState{
			AccountID: account.ID,
			State:     model.ParseAccountStatus(state),
		})
		if err != nil {
			return err
		}

		result = common.AccountInfo{
			AccountID: uint64(account.ID),
			Status:    string(status.State),
		}

		return nil
	})

	if err == nil {
		log.WithFields(logrus.Fields{
			"AccountID": result.AccountID,
			"Status":    result.Status,
		}).Debug("Account status updated")
	}

	return result, err
}

func OnAccountSetStatus(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountSetStatus")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
				"Status":    request.Status,
			})

			account, err := AccountSetStatus(ctx, request.AccountID, request.Status)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AccountSetStatus")
				return nil, internal.ErrInternalError
			}

			log.Info("Account status updated")

			// create & return response
			return &common.AccountInfo{
				AccountID: account.AccountID,
				Status:    account.Status,
			}, nil
		})
}
