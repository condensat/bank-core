// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"

	"github.com/condensat/bank-core/wallet/rpc"
)

func GetRawTransaction(ctx context.Context, rpcClient RpcClient, txID TransactionID) (Transaction, error) {
	var result Transaction
	err := callCommand(rpcClient, CmdGetRawTransaction, &result, txID)
	if err != nil {
		return "", rpc.ErrRpcError
	}

	return result, nil
}
