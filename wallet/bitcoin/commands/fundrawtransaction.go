// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func FundRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (FundedTransaction, error) {
	var result FundedTransaction
	err := callCommand(rpcClient, CmdFundRawTransaction, &result, hex)
	if err != nil {
		return FundedTransaction{}, err
	}
	return result, nil
}

func FundRawTransactionWithOptions(ctx context.Context, rpcClient RpcClient, hex Transaction, options FundRawTransactionOptions) (FundedTransaction, error) {
	var result FundedTransaction
	err := callCommand(rpcClient, CmdFundRawTransaction, &result, hex, &options)
	if err != nil {
		return FundedTransaction{}, err
	}
	return result, nil
}
