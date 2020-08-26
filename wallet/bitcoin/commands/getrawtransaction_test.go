// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestGetRawTransaction(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Errorf("Invalid RpcClient")
	}

	rpcClient := testRpcClient("bitcoin-testnet", 18332)
	if rpcClient == nil {
		t.Errorf("Invalid RpcClient")
		return
	}

	txID := TransactionID("4b8a545fd975e42021f5259afeda9799b677c5d753c4e551a80cbade18bb1753")
	rawTx := Transaction("010000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff450308a0110004449978590495932e310c59306259b4030000000000000a636b706f6f6c212f6d696e65642062792077656564636f646572206d6f6c69206b656b636f696e2fffffffff0277bc190b000000001976a91427f60a3b92e8a92149b18210457cc6bdc14057be88ac0000000000000000266a24aa21a9ed9851685d013d9a0271faea3784d5dc7b87304503cc4e9530cc5f2371bf311e900120000000000000000000000000000000000000000000000000000000000000000000000000")

	if len(txID) == 0 {
		t.Error("Invalid txID")
	}
	if len(rawTx) == 0 {
		t.Error("Invalid rawTx")
	}

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		txID      TransactionID
	}
	tests := []struct {
		name    string
		args    args
		want    Transaction
		wantErr bool
	}{
		// {"GetRawTransaction", args{ctx, rpcClient, txID}, rawTx, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRawTransaction(tt.args.ctx, tt.args.rpcClient, tt.args.txID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRawTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetRawTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
