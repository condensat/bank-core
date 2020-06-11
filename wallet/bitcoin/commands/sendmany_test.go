// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"testing"

	"github.com/condensat/bank-core/utils"
)

func TestSendManyWithFees(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Errorf("Invalid RpcClient")
	}

	rpcClient := testRpcClient("bitcoin-testnet", 18332)
	if rpcClient == nil {
		t.Logf("Invalid RpcClient")
		return
	}

	unspent, err := ListUnspent(ctx, rpcClient, nil)
	if err != nil {
		t.Logf("ListUnspent failed")
		return
	}
	if len(unspent) == 0 {
		t.Logf("ListUnspent empty")
		return
	}

	var totalAmount float64
	for _, info := range unspent {
		totalAmount += info.Amount
	}
	totalAmount = utils.ToFixed(totalAmount, 8) // satoshi precision

	t.Logf("totalAmount: %v", totalAmount)

	type args struct {
		ctx          context.Context
		rpcClient    RpcClient
		comment      string
		payinfo      []PayInfo
		fees         []Address
		estimateMode EstimateMode
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {"sendmany", args{ctx, rpcClient, "Shut up and take my money",
		// 	[]PayInfo{{Address: "tb1qg6vqcgg6423nx8smkg36xuydd2us06jfx6g9nn", Amount: totalAmount}},
		// 	[]Address{"tb1qg6vqcgg6423nx8smkg36xuydd2us06jfx6g9nn"},
		// 	EstimateModeEconomical,
		// }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SendManyWithFees(tt.args.ctx, tt.args.rpcClient, tt.args.comment, tt.args.payinfo, tt.args.fees, tt.args.estimateMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendManyWithFees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("SendManyWithFees() = %+v", got)
		})
	}
}
