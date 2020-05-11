// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"

	"github.com/condensat/bank-core/wallet/rpc"
)

func ListLockUnspent(ctx context.Context, rpcClient RpcClient) ([]UTXOInfo, error) {
	var list []UTXOInfo
	err := callCommand(rpcClient, CmdListLockUnspent, &list)
	if err != nil {
		return nil, rpc.ErrRpcError
	}

	return list, nil
}
