// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestImportAddress(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Errorf("Invalid RpcClient")
	}

	rpcClient := testRpcClient("bitcoin-testnet", 18332)
	if rpcClient == nil {
		t.Errorf("Invalid RpcClient")
		return
	}

	address := Address("tb1q29neqwhha0a94de7j7vewz4hkzwup5c75jz99u")

	if len(address) == 0 {
		t.Error("Invalid address")
	}

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		address   Address
		label     string
		reindex   bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {"ImportAddress", args{ctx, rpcClient, address, "test", false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ImportAddress(tt.args.ctx, tt.args.rpcClient, tt.args.address, tt.args.label, tt.args.reindex); (err != nil) != tt.wantErr {
				t.Errorf("ImportAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
