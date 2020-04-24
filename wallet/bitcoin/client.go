// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bitcoin

import (
	"context"
	"errors"
	"fmt"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/wallet/common"

	"github.com/btcsuite/btcd/chaincfg"
	rpc "github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
)

var (
	ErrInternalError  = errors.New("Internal Error")
	ErrRPCError       = errors.New("RPC Error")
	ErrInvalidAccount = errors.New("Invalid Account")
	ErrInvalidAddress = errors.New("Invalid Address format")
)

type BitcoinClient struct {
	conn   *rpc.ConnConfig
	client *rpc.Client
	params *chaincfg.Params
}

func paramsFromRPCPort(port int) *chaincfg.Params {
	params := &chaincfg.MainNetParams
	if port == 18332 {
		params = &chaincfg.TestNet3Params
	}
	return params
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
		params: paramsFromRPCPort(options.Port),
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

func (p *BitcoinClient) GetNewAddress(ctx context.Context, account string) (string, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.GetNewAddress")
	client := p.client
	if p.client == nil {
		return "", ErrInternalError
	}
	if len(account) == 0 {
		return "", ErrInvalidAccount
	}

	address, err := client.GetNewAddress(account)
	if err != nil {
		log.WithError(err).
			Error("GetNewAddress failed")
		return "", ErrRPCError
	}

	result := address.EncodeAddress()
	log.
		WithField("Address", result).
		Debug("Bitcoin RPC")

	return result, err
}

func (p *BitcoinClient) ListUnspent(ctx context.Context, minConf, maxConf int, addresses ...string) ([]common.AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.ListUnspent")

	client := p.client
	if p.client == nil {
		return nil, ErrInternalError
	}

	var filter []btcutil.Address
	for _, addr := range addresses {
		pubKey, err := btcutil.DecodeAddress(addr, p.params)
		if err != nil {
			log.WithError(err).
				WithField("Address", addr).
				Error("DecodeAddress failed")
			continue
		}
		filter = append(filter, pubKey)
	}

	if minConf > maxConf {
		minConf, maxConf = maxConf, minConf
	}
	list, err := client.ListUnspentMinMaxAddresses(minConf, maxConf, filter)
	if err != nil {
		log.WithError(err).
			Error("ListUnspentMinMaxAddresses failed")
		return nil, ErrRPCError
	}

	var result []common.AddressInfo
	for _, tx := range list {
		result = append(result, common.AddressInfo{
			Account:       tx.Account,
			Address:       tx.Address,
			TxID:          tx.TxID,
			Amount:        tx.Amount,
			Confirmations: tx.Confirmations,
			Spendable:     tx.Spendable,
		})
	}

	log.
		WithField("Count", len(list)).
		Debug("Bitcoin RPC")

	return result, nil
}
