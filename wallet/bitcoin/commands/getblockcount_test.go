// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestGetBlockCount(t *testing.T) {
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

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {"getblockcount", args{ctx, rpcClient}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockCount(tt.args.ctx, tt.args.rpcClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == 0 {
				t.Errorf("GetBlockCount() = %v", got)
			}

			t.Logf("GetBlockCount() = %v", got)
		})
	}
}
