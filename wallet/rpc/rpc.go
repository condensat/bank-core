// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rpc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ybbus/jsonrpc"
)

var (
	ErrRpcError = errors.New("Rpc Error")
)

type Client struct {
	Client jsonrpc.RPCClient
}

func New(options Options) *Client {
	// default values
	protocol := options.Protocol
	if len(protocol) == 0 {
		protocol = "http"
	}
	hostname := options.HostName
	if len(hostname) == 0 {
		hostname = "127.0.0.1"
	}
	port := options.Port
	if port == 0 {
		port = 4242
	}
	endpoint := options.Endpoint
	if len(endpoint) == 0 {
		endpoint = "/"
	}
	if strings.HasPrefix(endpoint, "/") {
		endpoint = strings.TrimLeft(endpoint, "/")
	}

	// format endpoint
	endpoint = fmt.Sprintf("%s://%s:%d/%s", protocol, hostname, port, endpoint)

	opts := options.rpcOption()
	return &Client{
		Client: jsonrpc.NewClientWithOpts(endpoint, &opts),
	}
}
