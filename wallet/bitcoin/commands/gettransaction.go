// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"

	"github.com/condensat/bank-core/wallet/rpc"
)

type GenericJson map[string]interface{}

func GetTransaction(ctx context.Context, rpcClient RpcClient, txID string) (TransactionInfo, error) {
	var obj GenericJson
	err := callCommand(rpcClient, CmdGetTransaction, &obj, txID)
	if err != nil {
		return TransactionInfo{}, rpc.ErrRpcError
	}

	return parseTransactionData(obj)
}

func parseTransactionData(obj GenericJson) (TransactionInfo, error) {
	data, err := json.Marshal(&obj)
	if err != nil {
		return TransactionInfo{}, err
	}

	result, err := parseTransactionInfo(data)
	if err != nil {
		return TransactionInfo{}, rpc.ErrRpcError
	}

	return result, nil
}

func parseTransactionInfo(data []byte) (TransactionInfo, error) {
	var info TransactionInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		// try to unmarshal elements
		info, err = parseElementsTransactionInfo(data)
		if err != nil {
			return TransactionInfo{}, err
		}
	}

	result := info
	if len(info.Address) == 0 && len(info.Details) == 1 {
		tx := info.Details[0]
		result = TransactionInfo{
			TxID:    info.TxID,
			Vout:    tx.Vout,
			Address: Address(tx.Address),

			Amount:        tx.Amount,
			AmountBlinder: tx.AmountBlinder,
			Asset:         tx.Asset,
			AssetBlinder:  tx.AssetBlinder,

			Confirmations: info.Confirmations,
		}
	}

	return result, err
}

type AmountMap map[string]float64

type ElementsTransactionInfo struct {
	Amount            AmountMap `json:"amount"`
	Confirmations     int64     `json:"confirmations"`
	Blockhash         string    `json:"blockhash"`
	Blockindex        int       `json:"blockindex"`
	Blocktime         int       `json:"blocktime"`
	TxID              string    `json:"txid"`
	Time              int       `json:"time"`
	TimeReceived      int       `json:"timereceived"`
	Bip125Replaceable string    `json:"bip125-replaceable"`
	Details           []struct {
		Address       string  `json:"address"`
		Category      string  `json:"category"`
		Amount        float64 `json:"amount"`
		AmountBlinder string  `json:"amountblinder"`
		Asset         string  `json:"asset"`
		AssetBlinder  string  `json:"assetblinder"`
		Label         string  `json:"label"`
		Vout          int     `json:"vout"`
	} `json:"details"`
}

func parseElementsTransactionInfo(data []byte) (TransactionInfo, error) {
	var info ElementsTransactionInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return TransactionInfo{}, err
	}

	if len(info.Details) != 1 {
		return TransactionInfo{}, err
	}

	// create TransactionInfo with first transaction detail
	tx := info.Details[0]
	return TransactionInfo{
		TxID:    info.TxID,
		Vout:    tx.Vout,
		Address: Address(tx.Address),

		Amount:        tx.Amount,
		AmountBlinder: tx.AmountBlinder,
		Asset:         tx.Asset,
		AssetBlinder:  tx.AssetBlinder,

		Confirmations: info.Confirmations,
	}, nil
}
