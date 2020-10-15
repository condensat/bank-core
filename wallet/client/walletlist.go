// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/wallet/common"

	"github.com/sirupsen/logrus"
)

func WalletList(ctx context.Context) ([]string, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.WalletList")

	var request common.WalletStatus
	var response common.WalletStatus
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.WalletListSubject, &request, &response)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return nil, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(response.Wallets),
	}).Debug("Wallet Info")

	var result []string
	for _, walletInfo := range response.Wallets {
		if len(walletInfo.Chain) == 0 {
			continue
		}
		result = append(result, walletInfo.Chain)
	}
	return result, nil
}
