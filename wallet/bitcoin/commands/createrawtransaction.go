// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"

	"github.com/condensat/bank-core/utils"
)

func CreateRawTransaction(ctx context.Context, rpcClient RpcClient, inputs []UTXOInfo, outputs []SpendInfo, assets []AssetInfo) (Transaction, error) {
	if inputs == nil {
		inputs = make([]UTXOInfo, 0)
	}

	// rpc args
	data := []interface{}{inputs}

	// gather same address outputs
	inputData := make(map[string]float64)
	for _, output := range outputs {
		if _, ok := inputData[output.Address]; !ok {
			inputData[output.Address] = 0.0
		}
		inputData[output.Address] += output.Amount
	}

	// Fix satoshi precision
	for address, totalAmount := range inputData {
		inputData[address] = utils.ToFixed(totalAmount, 8)
	}
	data = append(data, inputData, 0, false)

	// manage assets if provided
	if len(assets) > 0 {
		assetData := make(map[string]string)
		for _, asset := range assets {
			assetData[asset.Address] = asset.Asset
		}
		data = append(data, assetData)
	}

	var result Transaction
	err := callCommand(rpcClient, CmdCreateRawTransaction, &result, data...)
	if err != nil {
		return "", err
	}

	return result, nil
}
