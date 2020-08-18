// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func BlindRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (Transaction, error) {
	var result Transaction
	err := callCommand(rpcClient, CmdBlindRawTransaction, &result, hex)
	if err != nil {
		return Transaction(""), err
	}

	return result, nil
}
