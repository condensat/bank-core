// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
)

type CryptoStatus struct {
	Swap    SwapStatus    `json:"swap"`
	Reserve ReserveStatus `json:"reserve"`
}

func FetchCryptoStatus(ctx context.Context) (CryptoStatus, error) {
	swapStatus, err := FetchSwapStatus(ctx)
	if err != nil {
		return CryptoStatus{}, err
	}

	reserveStatus, err := FetchReserveStatus(ctx)
	if err != nil {
		return CryptoStatus{}, err
	}

	return CryptoStatus{
		Swap:    swapStatus,
		Reserve: reserveStatus,
	}, nil
}
