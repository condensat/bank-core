// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestImportPubKey(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Errorf("Invalid RpcClient")
	}

	rpcClient := testRpcClient("bitcoin-testnet", 18332)
	if rpcClient == nil {
		t.Errorf("Invalid RpcClient")
		return
	}

	pubKey := PubKey("0359732028d9cd03fa5316a69b5d6e81b978369bcfe1d7dc5637119f7ac5c9b210")

	if len(pubKey) == 0 {
		t.Error("Invalid pubKey")
	}

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		pubKey    PubKey
		label     string
		reindex   bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {"ImportPubKey", args{ctx, rpcClient, pubKey, "test", false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ImportPubKey(tt.args.ctx, tt.args.rpcClient, tt.args.pubKey, tt.args.label, tt.args.reindex); (err != nil) != tt.wantErr {
				t.Errorf("ImportPubKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
