// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestListLockUnspent(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Errorf("Invalid RpcClient")
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
		// {"ListLockUnspent", args{ctx, rpcClient}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListLockUnspent(tt.args.ctx, tt.args.rpcClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListLockUnspent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("ListLockUnspent() = %+v", got)
		})
	}
}
