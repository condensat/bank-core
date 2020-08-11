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

func ListUnspent(ctx context.Context, rpcClient RpcClient, filter []Address) ([]TransactionInfo, error) {
	return ListUnspentMinMaxAddresses(ctx, rpcClient, AddressInfoMinConfirmation, AddressInfoMaxConfirmation, filter)
}

func ListUnspentMinMaxAddresses(ctx context.Context, rpcClient RpcClient, minConf, maxConf int, filter []Address) ([]TransactionInfo, error) {
	list := make([]TransactionInfo, 0)
	err := callCommand(rpcClient, CmdListUnspent, &list, minConf, maxConf, filter)
	if err != nil {
		return nil, err
	}

	return list, nil
}
