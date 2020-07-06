// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/sirupsen/logrus"
)

func UserWithdrawList(ctx context.Context, userID uint64) (common.UserWithdraws, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.UserWithdrawList")

	// Database Query
	result := common.UserWithdraws{
		UserID: userID,
	}
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {
		withdraws, err := database.FindWithdrawByUser(db, model.UserID(userID))
		if err != nil {
			log.WithError(err).
				Error("AddWithdraw failed")
			return err
		}

		for _, w := range withdraws {
			wi, err := database.GetLastWithdrawInfo(db, w.ID)
			if err != nil {
				log.WithError(err).
					Error("GetLastWithdrawInfo failed")
				continue
			}

			wt, err := database.GetWithdrawTargetByWithdrawID(db, w.ID)
			if err != nil {
				log.WithError(err).
					Error("GetWithdrawTargetByWithdrawID failed")
				continue
			}

			data, err := wt.OnChainData()
			if err != nil {
				log.WithError(err).
					Error("AddWithdraw failed")
				continue
			}

			result.Withdraws = append(result.Withdraws, common.WithdrawInfo{
				WithdrawID: uint64(w.ID),
				Timestamp:  w.Timestamp,
				AccountID:  uint64(w.From),
				Amount:     float64(*w.Amount),
				Chain:      data.Chain,
				PublicKey:  data.PublicKey,
				Status:     string(wi.Status),
			})
		}

		return nil
	})

	if err != nil {
		return common.UserWithdraws{}, err
	}

	return result, nil
}

func OnUserWithdrawList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountTransferWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.UserWithdraws
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := UserWithdrawList(ctx, request.UserID)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"UserID": request.UserID,
					}).Errorf("Failed to UserWithdrawList")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}
