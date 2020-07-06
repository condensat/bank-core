// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func DecodeRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (RawTransaction, error) {
	var result RawTransaction
	err := callCommand(rpcClient, CmdDecodeRawTransaction, &result, hex)
	if err != nil {
		return RawTransaction{}, err
	}

	return result, nil
}
