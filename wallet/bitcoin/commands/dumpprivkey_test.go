// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"reflect"
	"testing"
)

func TestDumpPrivkey(t *testing.T) {
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

	const address Address = "tb1qwul7e5athct7m6xvunqmtyv8t4h3lu3a9pzhk0"
	const privkey Address = "cTphSFuckCXTiEnUMcawrQQbH76cwgSJQN52HQSxTaiHWwbN898E"

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		address   Address
	}
	tests := []struct {
		name    string
		args    args
		want    Address
		wantErr bool
	}{
		// {"dumpprivkey", args{ctx, rpcClient, address}, privkey, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DumpPrivkey(tt.args.ctx, tt.args.rpcClient, tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("DumpPrivkey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DumpPrivkey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDumpPrivkeys(t *testing.T) {
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

	const address Address = "tb1qwul7e5athct7m6xvunqmtyv8t4h3lu3a9pzhk0"
	const privkey Address = "cTphSFuckCXTiEnUMcawrQQbH76cwgSJQN52HQSxTaiHWwbN898E"

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		addresses []Address
	}
	tests := []struct {
		name    string
		args    args
		want    []Address
		wantErr bool
	}{
		// {"dumpprivkeys", args{ctx, rpcClient, []Address{address}}, []Address{privkey}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DumpPrivkeys(tt.args.ctx, tt.args.rpcClient, tt.args.addresses)
			if (err != nil) != tt.wantErr {
				t.Errorf("DumpPrivkeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DumpPrivkeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
