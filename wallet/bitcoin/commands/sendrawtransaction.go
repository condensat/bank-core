// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func SendRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (TxID, error) {
	var result TxID
	err := callCommand(rpcClient, CmdSendRawTransaction, &result, hex)
	if err != nil {
		return "", err
	}

	return result, nil
}
