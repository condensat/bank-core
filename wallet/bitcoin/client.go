// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bitcoin

import (
	"context"
	"errors"
	"fmt"

	"github.com/condensat/bank-core/logger"
	rpc "github.com/btcsuite/btcd/rpcclient"
)

var (
	ErrInternalError = errors.New("Internal Error")
	ErrRPCError      = errors.New("RPC Error")
)

type BitcoinClient struct {
	conn   *rpc.ConnConfig
	client *rpc.Client
}

func New(ctx context.Context, options BitcoinOptions) *BitcoinClient {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.New")
	connCfg := &rpc.ConnConfig{
		Host:         fmt.Sprintf("%s:%d", options.HostName, options.Port),
		User:         options.User,
		Pass:         options.Pass,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	client, err := rpc.New(connCfg, nil)
	if err != nil {
		log.WithError(err).
			Error("Failed to connect to bitcoin rpc server")
	}

	return &BitcoinClient{
		conn:   connCfg,
		client: client,
	}
}

func (p *BitcoinClient) GetBlockCount(ctx context.Context) (int64, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.GetBlockCount")
	client := p.client
	if p.client == nil {
		return 0, ErrInternalError
	}

	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.WithError(err).Error("GetBlockCount failed")
		return blockCount, ErrRPCError
	}

	log.
		WithField("BlockCount", blockCount).
		Debug("Bitcoin RPC")

	return blockCount, nil
}
