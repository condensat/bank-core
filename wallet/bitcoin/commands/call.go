// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

type RpcClient interface {
	CallFor(out interface{}, method string, params ...interface{}) error
}

func callCommand(rpcClient RpcClient, command Command, out interface{}, params ...interface{}) error {
	return rpcClient.CallFor(out, string(command), params...)
}
