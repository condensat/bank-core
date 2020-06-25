// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSignRawTransactionWithKey(t *testing.T) {
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
	var addreses = []Address{"cTphSFuckCXTiEnUMcawrQQbH76cwgSJQN52HQSxTaiHWwbN898E"}
	if len(addreses) == 0 {
		t.Errorf("Invalid addreses")
		return
	}

	var signed SignedTransaction
	_ = json.Unmarshal([]byte(mockSignedData), &signed)

	type args struct {
		ctx       context.Context
		rpcClient RpcClient
		hex       Transaction
		addreses  []Address
	}
	tests := []struct {
		name    string
		args    args
		want    SignedTransaction
		wantErr bool
	}{
		// {"signrawtransactionwithkey", args{ctx, rpcClient, hex, addreses}, signed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SignRawTransactionWithKey(tt.args.ctx, tt.args.rpcClient, tt.args.hex, tt.args.addreses)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignRawTransactionWithKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SignRawTransactionWithKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

const mockSignedData = `{
  "complete": true,
  "hex": "020000000001015f91840dd8aebc56d6f69025d5e3eb7e8a6bc568f9ff46d79b800c546c3b6d8e0000000000feffffff0280969800000000001600141d3e8b7739e16af543066df0e5b58fa27132b9e1f4495d05000000001600141e81b68331acb70c1adf50572afe6e26c5023944024730440220582e9f63f5286215f0f278bf39cf7e50607cab584bae892eeb7f3276a999c47e02205a5fdeca612a358dbb67f0c393b68002846ccbb155745ea27bbe507dde85156a01210282cacc06ba49fe6a906ceed4dfb1b8f1c7a607f8f3f2ee796f1e3bc7e7bcde1000000000"
}`
