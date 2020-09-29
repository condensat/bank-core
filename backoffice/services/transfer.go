// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
)

type DepositStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type BatchStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type WithdrawStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type TransferStatus struct {
	Deposit  DepositStatus  `json:"deposit"`
	Batch    BatchStatus    `json:"batch"`
	Withdraw WithdrawStatus `json:"withdraw"`
}

func FetchTransferStatus(ctx context.Context) (TransferStatus, error) {
	db := appcontext.Database(ctx)

	batchs, err := database.BatchsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	deposits, err := database.DepositsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	witdthdraws, err := database.WithdrawsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	return TransferStatus{
		Deposit: DepositStatus{
			Count:      deposits.Count,
			Processing: deposits.Active,
		},
		Batch: BatchStatus{
			Count:      batchs.Count,
			Processing: batchs.Active,
		},
		Withdraw: WithdrawStatus{
			Count:      witdthdraws.Count,
			Processing: witdthdraws.Active,
		},
	}, nil
}
