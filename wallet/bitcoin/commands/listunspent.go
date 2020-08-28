// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

const (
	AddressInfoMinConfirmation = 0
	AddressInfoMaxConfirmation = 6
)

type ListUnspentOption struct {
	MinimumAmount    float64 `json:"minimumAmount,omitempty"`
	MaximumAmount    float64 `json:"maximumAmount,omitempty"`
	MaximumCount     int     `json:"maximumCount,omitempty"`
	MinimumSumAmount float64 `json:"minimumSumAmount,omitempty"`
	Asset            string  `json:"asset,omitempty"`
}

func ListUnspent(ctx context.Context, rpcClient RpcClient, filter []Address) ([]TransactionInfo, error) {
	return ListUnspentMinMaxAddressesAndOptions(ctx, rpcClient, AddressInfoMinConfirmation, AddressInfoMaxConfirmation, filter, ListUnspentOption{})
}

func ListUnspentMinMaxAddressesAndOptions(ctx context.Context, rpcClient RpcClient, minConf, maxConf int, filter []Address, option ListUnspentOption) ([]TransactionInfo, error) {
	list := make([]TransactionInfo, 0)
	const includeUnsafe = true
	err := callCommand(rpcClient, CmdListUnspent, &list, minConf, maxConf, filter, includeUnsafe, option)
	if err != nil {
		return nil, err
	}

	return list, nil
}
