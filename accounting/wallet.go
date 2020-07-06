// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"

	"github.com/condensat/bank-core/logger"
	"github.com/sirupsen/logrus"
)

type ChainOutput struct {
	PublicKey string
	Amount    float64
}

func SentWalletBatchRequest(ctx context.Context, chain string, outputs []ChainOutput) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.SentWalletBatchRequest")

	log.WithFields(logrus.Fields{
		"Chain":   chain,
		"Outputs": outputs,
	}).Debug("Sending batch to wallet")
	return nil
}
