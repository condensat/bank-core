// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
)

type SwapStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

func FetchSwapStatus(ctx context.Context) (SwapStatus, error) {
	db := appcontext.Database(ctx)

	swaps, err := database.SwapssInfos(db)
	if err != nil {
		return SwapStatus{}, err
	}

	return SwapStatus{
		Count:      swaps.Count,
		Processing: swaps.Active,
	}, nil
}
