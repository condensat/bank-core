// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func GetAddressInfo(ctx context.Context, rpcClient RpcClient, address Address) (AddressInfo, error) {
	var result AddressInfo
	err := callCommand(rpcClient, CmdGetAddressInfo, &result, address)
	if err != nil {
		return AddressInfo{}, err
	}

	return result, nil
}
