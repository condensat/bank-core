// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestLockUnspent(t *testing.T) {
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

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		unlock    bool
		utxos     []UTXOInfo
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// {"LockUnspent", args{ctx, rpcClient, unlock, utxos}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LockUnspent(tt.args.ctx, tt.args.rpcClient, tt.args.unlock, tt.args.utxos)
			if (err != nil) != tt.wantErr {
				t.Errorf("LockUnspent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LockUnspent() = %v, want %v", got, tt.want)
			}
		})
	}
}
