// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestListUnspent(t *testing.T) {
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
		filter    []Address
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {"listunspent", args{ctx, rpcClient, nil}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListUnspent(tt.args.ctx, tt.args.rpcClient, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUnspentAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("ListUnspent() = %v", got)
		})
	}
}
