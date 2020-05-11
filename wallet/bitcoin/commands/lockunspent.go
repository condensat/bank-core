// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"

	"github.com/condensat/bank-core/wallet/rpc"
)

func LockUnspent(ctx context.Context, rpcClient RpcClient, unlock bool, utxos []UTXOInfo) (bool, error) {
	var success bool
	err := callCommand(rpcClient, CmdLockUnspent, &success, unlock, utxos)
	if err != nil {
		return false, rpc.ErrRpcError
	}

	return success, nil
}
