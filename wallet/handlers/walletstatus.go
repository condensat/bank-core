// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func WalletStatus(ctx context.Context, status common.WalletStatus) (common.WalletStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.WalletStatus")
	var result common.WalletStatus

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return result, ErrInternalError
	}

	// Todo: list all available chains
	var chains []string
	for _, chain := range chains {
		walletInfo, err := chainHandler.WalletInfo(ctx, chain)
		if err != nil {
			log.WithError(err).
				WithField("Chain", chain).
				Warning("WalletInfo Failed")
			continue
		}
		result.Wallets = append(result.Wallets, walletInfo)
	}

	return result, nil
}

func OnWalletStatus(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnWalletStatus")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.WalletStatus
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			status, err := WalletStatus(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to WalletStatus")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &status, nil
		})
}
