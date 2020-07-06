// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"
)

func TestSendRawTransaction(t *testing.T) {
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

	const hex Transaction = "02000000015f91840dd8aebc56d6f69025d5e3eb7e8a6bc568f9ff46d79b800c546c3b6d8e0000000000feffffff0280969800000000001600141d3e8b7739e16af543066df0e5b58fa27132b9e1f4495d05000000001600141e81b68331acb70c1adf50572afe6e26c502394400000000"
	const txID TxID = "1ef8baa0b133e3d16ee7e12c08682be7626c5b9c98a0b956dd1f292c8410ced0"

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		hex       Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    TxID
		wantErr bool
	}{
		// {"sendrawtransaction", args{ctx, rpcClient, hex}, txID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SendRawTransaction(tt.args.ctx, tt.args.rpcClient, tt.args.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendRawTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SendRawTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
