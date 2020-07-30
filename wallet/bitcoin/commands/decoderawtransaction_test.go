// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestDecodeRawTransaction(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Errorf("Invalid RpcClient")
		return
	}

	rpcClient := testRpcClient("bitcoin-testnet", 18332)
	if rpcClient == nil {
		t.Errorf("Invalid RpcClient")
		return
	}

	var raw RawTransaction
	_ = json.Unmarshal([]byte(mockRawTransactionData), &raw)
	const hex Transaction = "02000000000180969800000000001600141d3e8b7739e16af543066df0e5b58fa27132b9e100000000"
	if len(hex) == 0 {
		t.Error("Invalid hex")
	}

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		hex       Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    RawTransaction
		wantErr bool
	}{
		// {"decoderawtransaction", args{ctx, rpcClient, hex}, raw, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeRawTransaction(tt.args.ctx, tt.args.rpcClient, tt.args.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeRawTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeRawTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

const mockRawTransactionData = `{
	"hash": "7a7a8c885d77431c8699a79f56faaf3411fa5fb17aa0d98f27b92e8fcaa050a2",
	"locktime": 0,
	"size": 41,
	"txid": "7a7a8c885d77431c8699a79f56faaf3411fa5fb17aa0d98f27b92e8fcaa050a2",
	"version": 2,
	"vin": [],
	"vout": [
		{
			"n": 0,
			"scriptPubKey": {
				"addresses": [
					"tb1qr5lgkaeeu9402scxdhcwtdv05fcn9w0pq45usg"
				],
				"asm": "0 1d3e8b7739e16af543066df0e5b58fa27132b9e1",
				"hex": "00141d3e8b7739e16af543066df0e5b58fa27132b9e1",
				"reqSigs": 1,
				"type": "witness_v0_keyhash"
			},
			"value": 0.1
		}
	],
	"vsize": 41,
	"weight": 164
}`
