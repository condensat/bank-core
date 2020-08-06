// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"context"
	"sync"
)

const (
	ChainClientKey   = "Key.ChainClientKey"
	SsmClientKey     = "Key.SsmClientKey"
	SsmDeviceInfoKey = "Key.SsmDeviceInfoKey"
	CryptoModeKey    = "Key.CryptoModeKey"
)

func CryptoModeContext(ctx context.Context, mode CryptoMode) context.Context {
	return context.WithValue(ctx, CryptoModeKey, mode)
}

func CryptoModeFromContext(ctx context.Context) CryptoMode {
	if mode, ok := ctx.Value(CryptoModeKey).(CryptoMode); ok {
		return mode
	}
	return CryptoModeBitcoinCore
}

// Chains

func ChainClientContext(ctx context.Context, chain string, client ChainClient) context.Context {
	// check valid client
	if client == nil {
		// NOOP
		return ctx
	}
	// check if client is registered
	if client := ChainClientFromContext(ctx, chain); client != nil {
		// NOOP
		return ctx
	}

	// check if multiChainClient is presnet in context
	switch chains := ctx.Value(ChainClientKey).(type) {

	case *multiChainClient:
		// add client if not found
		if chains.Client(chain) == nil {
			chains.Add(chain, client)
		}
		return ctx

	default:
		// create pool
		ctx := context.WithValue(ctx, ChainClientKey, &multiChainClient{
			clients: make(map[string]ChainClient),
		})

		// add client to pool
		return ChainClientContext(ctx, chain, client)
	}
}

func ChainClientFromContext(ctx context.Context, chain string) ChainClient {
	switch chains := ctx.Value(ChainClientKey).(type) {
	case *multiChainClient:

		// return client form pool (can be null)
		return chains.Client(chain)

	default:
		return nil
	}
}

// Chainclient pool

type multiChainClient struct {
	sync.Mutex
	clients map[string]ChainClient
}

func (p *multiChainClient) Add(chain string, client ChainClient) {
	p.Lock()
	defer p.Unlock()

	p.clients[chain] = client
}

func (p *multiChainClient) Client(chain string) ChainClient {
	p.Lock()
	defer p.Unlock()

	client := p.clients[chain]
	return client
}

// Ssms

func SsmClientContext(ctx context.Context, device string, client SsmClient) context.Context {
	// check valid client
	if client == nil {
		// NOOP
		return ctx
	}
	// check if client is registered
	if client := ChainClientFromContext(ctx, device); client != nil {
		// NOOP
		return ctx
	}

	// check if multiSsmClient is presnet in context
	switch ssms := ctx.Value(SsmClientKey).(type) {

	case *multiSsmClient:
		// add client if not found
		if ssms.Client(device) == nil {
			ssms.Add(device, client)
		}
		return ctx

	default:
		// create pool
		ctx := context.WithValue(ctx, SsmClientKey, &multiSsmClient{
			clients: make(map[string]SsmClient),
		})

		// add client to pool
		return SsmClientContext(ctx, device, client)
	}
}

func SsmClientFromContext(ctx context.Context, device string) SsmClient {
	switch ssms := ctx.Value(SsmClientKey).(type) {
	case *multiSsmClient:

		// return client form pool (can be null)
		return ssms.Client(device)

	default:
		return nil
	}
}

// SsmClient pool
type multiSsmClient struct {
	sync.Mutex
	clients map[string]SsmClient
}

func (p *multiSsmClient) Add(device string, client SsmClient) {
	p.Lock()
	defer p.Unlock()

	p.clients[device] = client
}

func (p *multiSsmClient) Client(device string) SsmClient {
	p.Lock()
	defer p.Unlock()

	client := p.clients[device]
	return client
}

func SsmDeviceInfoContext(ctx context.Context, info SsmDeviceInfo) context.Context {
	return context.WithValue(ctx, CryptoModeKey, info)
}

func SsmDeviceInfoFromContext(ctx context.Context) SsmDeviceInfo {
	if info, ok := ctx.Value(SsmDeviceInfoKey).(SsmDeviceInfo); ok {
		return info
	}
	return nil
}
