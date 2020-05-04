// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"

	"github.com/condensat/bank-core/wallet/rpc"
)

const (
	AddressTypeBech32 = "bech32"
)

func GetNewAddress(ctx context.Context, rpcClient RpcClient, label, addressType string) (Address, error) {
	return GetNewAddressWithType(ctx, rpcClient, label, AddressTypeBech32)
}

func GetNewAddressWithType(ctx context.Context, rpcClient RpcClient, label, addressType string) (Address, error) {
	if rpcClient == nil {
		return "", ErrInvalidRPCClient
	}

	var address Address
	err := callCommand(rpcClient, CmdGetNewAddress, &address, label, addressType)
	if err != nil {
		return "", rpc.ErrRpcError
	}

	return address, nil
}
