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
			got, err := GetTransaction(tt.args.ctx, tt.args.rpcClient, tt.args.txID, true)
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

	mockElementsSwap := `{
		"amount": {
			"0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a": 0.00001400,
			"bitcoin": 0.00000571,
			"ce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2": -0.00001000
		},
		"fee": {
			"bitcoin": -0.00001143
		},
		"confirmations": 163,
		"blockhash": "36c28a979e7dd83b29fdb6bde66e349a34b6668a459f4acbf2469b3d9f8acb64",
		"blockindex": 1,
		"blocktime": 1589548436,
		"txid": "4f9c0a5c5245bf73201d0293a8930763d7586ddfa8aa939a175882b121402a27",
		"walletconflicts": [
		],
		"time": 1589548429,
		"timereceived": 1589548429,
		"bip125-replaceable": "no",
		"details": [
			{
				"address": "GiWHkKfUby65ZKtAcsXtgMpyLuMZnB1kAe",
				"category": "send",
				"amount": -0.00001000,
				"amountblinder": "1b2c8e14041c95d5c1e85b8bb923a0e6a041a018be02478d49808dfed2187dad",
				"asset": "ce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2",
				"assetblinder": "bb426cd1f8aed53b55b30377354f6d2e601e27e07c483fef420caa53fab21e60",
				"vout": 2,
				"fee": 0.00001143,
				"abandoned": false
			},
			{
				"address": "ex1qunmd9vf46l8ytrg3qq95paq5y402wdxqn7r8z9",
				"category": "send",
				"amount": -0.00009429,
				"amountblinder": "cc5a001d325368861ec3549b8680c9277940d2b8a858d6e05349ab28bf9c2eda",
				"asset": "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
				"assetblinder": "c9e09c143888373f37facd0e8837ae75f34df7e4ae9164ade4f3482e3915be86",
				"vout": 3,
				"fee": 0.00001143,
				"abandoned": false
			},
			{
				"address": "ex1qaxkxqgvdh37547leyjschhy2cx4wlm46pwplcr",
				"category": "send",
				"amount": -0.00001400,
				"amountblinder": "da73530d914e97f94bbaaf5b4d26c0138a1289c3f2600ec865cd8c95da28cc1b",
				"asset": "0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a",
				"assetblinder": "9b3dd8e79249458bfa95b08890a92fa822bf195367d1dc55806aca68fe123fe2",
				"label": "3ke2CqaPZHYgEM54EZW7P",
				"vout": 4,
				"fee": 0.00001143,
				"abandoned": false
			},
			{
				"address": "ex1qpwjpsf3v9crn94wanhgkz2cydv343nemu0ea64",
				"category": "send",
				"amount": -0.00997200,
				"amountblinder": "234ddd2018bc21ca91c81e0b571ffaa478176909a4f01c62de38159f9158c673",
				"asset": "0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a",
				"assetblinder": "1de4b87814bf95d524989ae41f37698fb8fb2bf886b8aa7b554dea97743cd022",
				"vout": 5,
				"fee": 0.00001143,
				"abandoned": false
			},
			{
				"address": "ex1qaxkxqgvdh37547leyjschhy2cx4wlm46pwplcr",
				"category": "receive",
				"amount": 0.00001400,
				"amountblinder": "da73530d914e97f94bbaaf5b4d26c0138a1289c3f2600ec865cd8c95da28cc1b",
				"asset": "0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a",
				"assetblinder": "9b3dd8e79249458bfa95b08890a92fa822bf195367d1dc55806aca68fe123fe2",
				"label": "3ke2CqaPZHYgEM54EZW7P",
				"vout": 4
			}
		],
		"hex": ""
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
	var objElementsSwap GenericJson
	err = json.Unmarshal([]byte(mockElementsSwap), &objElementsSwap)
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
		{"ParseElementsSwap", args{objElementsSwap}, false},
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
