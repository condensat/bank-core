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

func WalletList(ctx context.Context, wallets []common.WalletInfo) ([]string, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.WalletList")

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return nil, ErrInternalError
	}

	var result []string
	chains := chainHandler.ListChains(ctx)
	// return all chains if no chains specified in requests
	if len(wallets) == 0 {
		result = chains
	}

	// select only requested chains
	for _, wallet := range wallets {
		for _, chain := range chains {
			if chain != wallet.Chain {
				continue
			}
			result = append(result, chain)
		}
	}

	return result, nil
}

func OnWalletList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnWalletList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.WalletStatus
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			chains, err := WalletList(ctx, request.Wallets)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to WalletList")
				return nil, cache.ErrInternalError
			}

			// create & return response
			var status common.WalletStatus
			for _, chain := range chains {
				status.Wallets = append(status.Wallets, common.WalletInfo{Chain: chain})
			}
			return &status, nil
		})
}
