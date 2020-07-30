// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestCreateRawTransaction(t *testing.T) {
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

	var outputs []SpendInfo
	outputs = append(outputs, SpendInfo{
		Address: "tb1qr5lgkaeeu9402scxdhcwtdv05fcn9w0pq45usg",
		Amount:  0.1,
	})

	if len(outputs) == 0 {
		t.Error("Invalid output")
	}

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		inputs    []UTXOInfo
		outputs   []SpendInfo
	}
	tests := []struct {
		name    string
		args    args
		want    Transaction
		wantErr bool
	}{
		// {"createrawtransaction", args{ctx, rpcClient, nil, outputs}, "02000000000180969800000000001600141d3e8b7739e16af543066df0e5b58fa27132b9e100000000", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateRawTransaction(tt.args.ctx, tt.args.rpcClient, tt.args.inputs, tt.args.outputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRawTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateRawTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
