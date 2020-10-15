// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/rpc"
)

func testRpcClient(hostname string, port int) RpcClient {
	return rpc.New(rpc.Options{
		ServerOptions: common.ServerOptions{Protocol: "http", HostName: hostname, Port: port},
		User:          "condensat",
		Password:      "condensat",
	}).Client
}
