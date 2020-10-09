// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"
	"errors"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/wallet/common"

	"github.com/sirupsen/logrus"
)

const (
	WalletStatusWildcard = "*"
)

var (
	ErrInvalidChain = errors.New("Invalid Chain")
)

func WalletStatus(ctx context.Context, chain string) (common.WalletStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.WalletStatus")

	if len(chain) == 0 {
		return common.WalletStatus{}, ErrInvalidChain
	}

	var request common.WalletStatus
	if chain != WalletStatusWildcard {
		request.Wallets = append(request.Wallets, common.WalletInfo{
			Chain: chain,
		})
	}
	var result common.WalletStatus
	err := messaging.RequestMessage(ctx, common.WalletStatusSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.WalletStatus{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(result.Wallets),
	}).Debug("Wallet Info")

	return result, nil
}
