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

func CryptoAddressNextDeposit(ctx context.Context, chain string, accountID uint64) (common.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.CryptoAddressNextDeposit")

	request := common.CryptoAddress{
		Chain:     chain,
		AccountID: accountID,
	}

	var result common.CryptoAddress
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CryptoAddressNextDepositSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CryptoAddress{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"CryptoAddressID": result.CryptoAddressID,
		"Chain":           result.Chain,
		"AccountID":       result.AccountID,
		"PublicAddress":   result.PublicAddress,
	}).Debug("Next Deposit Address")

	return result, nil
}
