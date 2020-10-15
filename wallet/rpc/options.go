// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/base64"
	"fmt"

	"github.com/condensat/bank-core/wallet/common"
	"github.com/ybbus/jsonrpc"
)

type Options struct {
	common.ServerOptions
	User     string
	Password string

	Endpoint string
}

func (p *Options) rpcOption() jsonrpc.RPCClientOpts {
	var options jsonrpc.RPCClientOpts
	if len(p.User) > 0 {
		basic := fmt.Sprintf("%s:%s", p.User, p.Password)
		options.CustomHeaders = map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(basic)),
		}
	}
	return options
}
