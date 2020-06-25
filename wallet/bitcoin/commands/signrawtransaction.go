// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func SignRawTransactionWithKey(ctx context.Context, rpcClient RpcClient, hex Transaction, addreses []Address) (SignedTransaction, error) {
	var result SignedTransaction
	err := callCommand(rpcClient, CmdSignRawTransactionWithKey, &result, hex, addreses)
	if err != nil {
		return SignedTransaction{}, err
	}

	return result, nil
}

func SignRawTransactionWithWallet(ctx context.Context, rpcClient RpcClient, hex Transaction) (SignedTransaction, error) {
	var result SignedTransaction
	err := callCommand(rpcClient, CmdSignRawTransactionWithWallet, &result, hex)
	if err != nil {
		return SignedTransaction{}, err
	}

	return result, nil
}
