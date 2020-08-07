// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ssm

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"

	"github.com/condensat/bank-core"
	"github.com/ybbus/jsonrpc"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/wallet/common"
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

func NewWithTorEndpoint(ctx context.Context, torProxy, endpoint string) *SsmClient {
	proxyURL, err := url.Parse(torProxy)
	if err != nil {
		panic(err)
	}

	return &SsmClient{
		client: &rpc.Client{
			Client: jsonrpc.NewClientWithOpts(endpoint, &jsonrpc.RPCClientOpts{
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						Proxy: http.ProxyURL(proxyURL),
					},
				},
			}),
		},
	}
}

func (p *SsmClient) NewAddress(ctx context.Context, ssmPath commands.SsmPath) (common.SsmAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "ssm.NewAddress")

	client := p.client
	if p.client == nil {
		return common.SsmAddress{}, ErrInternalError
	}

	result, err := commands.NewAddress(ctx, client.Client, ssmPath.Chain, ssmPath.Fingerprint, ssmPath.Path)
	if err != nil {
		log.WithError(err).Error("NewAddress failed")
		return common.SsmAddress{}, ErrRPCError
	}

	log.
		WithField("Chain", result.Chain).
		WithField("Address", result.Address).
		WithField("PubKey", result.PubKey).
		WithField("BlindingKey", result.BlindingKey).
		Debug("SSM RPC")

	return common.SsmAddress{
		Chain:       result.Chain,
		Address:     result.Address,
		PubKey:      result.PubKey,
		BlindingKey: result.BlindingKey,
	}, nil
}

func (p *SsmClient) SignTx(ctx context.Context, chain, inputransaction string, inputs ...commands.SignTxInputs) (string, error) {
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

type SsmDeviceInfo struct {
	sync.Mutex
	info map[common.SsmChain]common.SsmFingerprint
}

func NewDeviceInfo(ctx context.Context) *SsmDeviceInfo {
	return &SsmDeviceInfo{
		info: make(map[common.SsmChain]common.SsmFingerprint),
	}
}

func (p *SsmDeviceInfo) Add(ctx context.Context, chain common.SsmChain, fingerprint common.SsmFingerprint) error {
	p.Lock()
	defer p.Unlock()

	if len(chain) == 0 {
		return errors.New("Invalid chain")
	}
	if len(fingerprint) == 0 {
		return errors.New("Invalid fingerprint")
	}
	if _, ok := p.info[chain]; ok {
		return errors.New("Chain fingerprint exists")
	}

	p.info[chain] = fingerprint

	return nil
}

func (p *SsmDeviceInfo) Fingerprint(ctx context.Context, chain common.SsmChain) (common.SsmFingerprint, error) {
	p.Lock()
	defer p.Unlock()

	if len(chain) == 0 {
		return "", errors.New("Invalid chain")
	}

	result, ok := p.info[chain]
	if !ok {
		return "", errors.New("Fingerprint not found")
	}

	return result, nil
}
