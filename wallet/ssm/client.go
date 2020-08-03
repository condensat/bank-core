// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ssm

import (
	"context"
	"errors"
	"sync"

	"github.com/condensat/bank-core"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/wallet/rpc"
	"github.com/condensat/bank-core/wallet/ssm/commands"
)

var (
	ErrInternalError    = errors.New("Internal Error")
	ErrRPCError         = errors.New("RPC Error")
	ErrInvalidAccount   = errors.New("Invalid Account")
	ErrInvalidAddress   = errors.New("Invalid Address format")
	ErrLockUnspentFails = errors.New("LockUnpent Failed")
)

type SsmClient struct {
	sync.Mutex // mutex to change params while RPC

	client *rpc.Client
}

func New(ctx context.Context, options SsmOptions) *SsmClient {
	client := rpc.New(rpc.Options{
		ServerOptions: bank.ServerOptions{Protocol: "http", HostName: options.HostName, Port: options.Port},
		User:          options.User,
		Password:      options.Pass,
	})

	return &SsmClient{
		client: client,
	}
}

func (p *SsmClient) NewAddress(ctx context.Context, chain, fingerprint, path string) (string, error) {
	log := logger.Logger(ctx).WithField("Method", "ssm.NewAddress")

	client := p.client
	if p.client == nil {
		return "", ErrInternalError
	}

	result, err := commands.NewAddress(ctx, client.Client, chain, fingerprint, path)
	if err != nil {
		log.WithError(err).Error("NewAddress failed")
		return "", ErrRPCError
	}

	log.
		WithField("Chain", result.Chain).
		WithField("Address", result.Address).
		Debug("SSM RPC")

	return result.Address, nil
}

func (p *SsmClient) SignTx(ctx context.Context, chain, inputransaction string, inputs []commands.SignTxInputs) (string, error) {
	log := logger.Logger(ctx).WithField("Method", "ssm.SignTx")

	client := p.client
	if p.client == nil {
		return "", ErrInternalError
	}

	result, err := commands.SignTx(ctx, client.Client, chain, inputransaction, inputs...)
	if err != nil {
		log.WithError(err).Error("SignTx failed")
		return "", ErrRPCError
	}

	log.
		WithField("Chain", result.Chain).
		WithField("SignedTx", result.SignedTx).
		Debug("SSM RPC")

	return result.SignedTx, nil
}
