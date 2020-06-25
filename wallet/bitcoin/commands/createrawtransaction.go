// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func CreateRawTransaction(ctx context.Context, rpcClient RpcClient, inputs []UTXOInfo, outputs []SpendInfo) (Transaction, error) {
	if inputs == nil {
		inputs = make([]UTXOInfo, 0)
	}

	data := make(map[string]float64)
	for _, output := range outputs {
		data[output.Address] = output.Amount
	}
	var result Transaction
	err := callCommand(rpcClient, CmdCreateRawTransaction, &result, inputs, data)
	if err != nil {
		return "", err
	}

	return result, nil
}
