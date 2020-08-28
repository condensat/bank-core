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

func TestFundRawTransaction(t *testing.T) {
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

	var raw FundedTransaction
	_ = json.Unmarshal([]byte(mockFundedTransaction), &raw)
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
		want    FundedTransaction
		wantErr bool
	}{
		// {"fundrawtransaction", args{ctx, rpcClient, hex}, raw, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FundRawTransaction(tt.args.ctx, tt.args.rpcClient, tt.args.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("FundRawTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// remove trailing time related output
			got.Hex = got.Hex[:len(got.Hex)-134]
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FundRawTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

const mockFundedTransaction = `{
  "changepos": 0,
  "fee": 1.41e-06,
	"hex": "02000000015f91840dd8aebc56d6f69025d5e3eb7e8a6bc568f9ff46d79b800c546c3b6d8e0000000000feffffff"
}`
