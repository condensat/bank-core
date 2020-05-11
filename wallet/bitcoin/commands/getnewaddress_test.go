// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"strings"
	"testing"
)

func TestGetNewAddress(t *testing.T) {
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
		ctx         context.Context
		rpcClient   RpcClient
		label       string
		addressType string
	}
	tests := []struct {
		name       string
		args       args
		wantPrefix string
		wantErr    bool
	}{
		// {"getnewaddress", args{ctx, rpcClient, "test", AddressTypeBech32}, "tb1", false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNewAddress(tt.args.ctx, tt.args.rpcClient, tt.args.label, tt.args.addressType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNewAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.HasPrefix(string(got), tt.wantPrefix) {
				t.Errorf("GetNewAddress() = %v, wrong prefix, want %v", got, tt.wantPrefix)
			}

			t.Logf("GetNewAddress() = %v", got)
		})
	}
}
