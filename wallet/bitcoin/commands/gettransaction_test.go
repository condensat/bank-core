// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetTransaction(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Errorf("Invalid RpcClient")
	}

	rpcClient := testRpcClient("bitcoin-testnet", 18332)
	if rpcClient == nil {
		t.Errorf("Invalid RpcClient")
		return
	}

	var utxos []UTXOInfo
	{
		unspent, err := ListUnspent(ctx, rpcClient, nil)
		if err != nil {
			t.Logf("ListUnspent failed")
			return
		}
		for _, tx := range unspent {
			utxos = append(utxos, UTXOInfo{
				TxID: tx.TxID,
				Vout: tx.Vout,
			})
		}
	}

	unlock := len(utxos) == 0

	// try to unlock
	if unlock {
		var err error
		utxos, err = ListLockUnspent(ctx, rpcClient)
		if err != nil {
			t.Errorf("ListLockUnspent failed")
			return
		}
	}

	if len(utxos) == 0 {
		t.Logf("Empty utxos")
		return
	}

	utxo := utxos[0]
	if len(utxo.TxID) == 0 {
		t.Logf("Invalid TxID")
		return
	}

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		txID      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {"GetTransaction", args{ctx, rpcClient, utxo.TxID}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTransaction(tt.args.ctx, tt.args.rpcClient, tt.args.txID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("GetTransaction() = %+v", got)

		})
	}
}

func Test_parseTransactionData(t *testing.T) {

	mockBtc := `{
    "amount": 0.00560775,
    "confirmations": 22,
    "blockhash": "000000004305a0356bba2f9ddc876dd34627fc0ae3bef98c3d70b9710ff4e953",
    "blockindex": 146,
    "blocktime": 1588832282,
    "txid": "f5990d0d6e51e30161c84feda24a4505539fe61dfa909d04c0d14d04790f8d1c",
    "walletconflicts": [],
    "time": 1588832195,
    "timereceived": 1588832195,
    "bip125-replaceable": "no",
    "details": [
      {
        "address": "tb1qzhut9p8zdnlsz9ac93vn7gp8xsysklxg52m2h9",
        "category": "receive",
        "amount": 0.00560775,
        "label": "3ke2CqaPZHYgEM54EZVaX",
        "vout": 1
      }
    ],
    "hex": "aaaa"
  }`

	mockElements := `{
		"amount": {
			"bd4f2425ee67adcf93a3c769416092b55cd649f0736ee95af035494a7015e531": 0.01,
			"bitcoin": 0
		},
		"confirmations": 5,
		"blockhash": "64e79b988b4cd09a6e42b01a4a20f63bfdca6318398ad761d46c451b161f7782",
		"blockindex": 1,
		"blocktime": 1588833417,
		"txid": "53c4dc96a79af930f5ec0ccd76a79a731bf06419685749fe8eeaca6137a7c85c",
		"walletconflicts": [],
		"time": 1588833358,
		"timereceived": 1588833358,
		"bip125-replaceable": "no",
		"details": [
			{
				"address": "ex1qs3h7x5l4tyqdeewe47vf48luum5uwd9wt4ayk8",
				"category": "receive",
				"amount": 0.01,
				"amountblinder": "b7578b0edeb9010f3df75e0a76623f4c241671c7e59ca9fb5c6dd687319daa51",
				"asset": "bd4f2425ee67adcf93a3c769416092b55cd649f0736ee95af035494a7015e531",
				"assetblinder": "cc43e1a92bbabcdd1b6378f85a15fc2fa6362351f206daec542673da91d35996",
				"label": "3ke2CqaPZHYgEM54EZVaW",
				"vout": 0
			}
		],
		"hex": "aaa"
	}`

	var objBtc GenericJson
	err := json.Unmarshal([]byte(mockBtc), &objBtc)
	if err != nil {
		t.Errorf("Invalid mock data: %s", err)
		return
	}
	var objElements GenericJson
	err = json.Unmarshal([]byte(mockElements), &objElements)
	if err != nil {
		t.Errorf("Invalid mock data: %s", err)
		return
	}

	type args struct {
		obj GenericJson
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ParseBtc", args{objBtc}, false},
		{"ParseElements", args{objElements}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTransactionData(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTransactionData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got.Address) == 0 {
				t.Errorf("parseTransactionData() Invalid Address %+v", got)
				return
			}
			t.Logf("parseTransactionData: %+v", got)
		})
	}
}
