// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bitcoin

import (
	"context"
	"errors"
	"sync"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/rpc"

	"github.com/condensat/bank-core/wallet/bitcoin/commands"

	"github.com/sirupsen/logrus"
)

var (
	ErrInternalError  = errors.New("Internal Error")
	ErrRPCError       = errors.New("RPC Error")
	ErrInvalidAccount = errors.New("Invalid Account")
	ErrInvalidAddress = errors.New("Invalid Address format")
)

const (
	AddressTypeBech32 = "bech32"
)

type BitcoinClient struct {
	sync.Mutex // mutex to change params while RPC

	client *rpc.Client
}

func New(ctx context.Context, options BitcoinOptions) *BitcoinClient {
	client := rpc.New(rpc.Options{
		ServerOptions: bank.ServerOptions{Protocol: "http", HostName: options.HostName, Port: options.Port},
		User:          options.User,
		Password:      options.Pass,
	})

	return &BitcoinClient{
		client: client,
	}
}

func (p *BitcoinClient) GetBlockCount(ctx context.Context) (int64, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.GetBlockCount")

	client := p.client
	if p.client == nil {
		return 0, ErrInternalError
	}

	blockCount, err := commands.GetBlockCount(ctx, client.Client)
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

	result, err := commands.GetNewAddressWithType(ctx, client.Client, account, AddressTypeBech32)
	if err != nil {
		log.WithError(err).
			Error("GetNewAddress failed")
		return "", ErrRPCError
	}

	log.
		WithFields(logrus.Fields{
			"Account": account,
			"Address": result,
			"Type":    AddressTypeBech32,
		}).Debug("Bitcoin RPC")

	return string(result), nil
}

func (p *BitcoinClient) GetAddressInfo(ctx context.Context, address string) (common.AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.GetNewAddress")

	client := p.client
	if p.client == nil {
		return common.AddressInfo{}, ErrInternalError
	}
	if len(address) == 0 {
		return common.AddressInfo{}, ErrInvalidAddress
	}

	info, err := commands.GetAddressInfo(ctx, client.Client, commands.Address(address))
	if err != nil {
		log.WithError(err).
			Error("GetAddressInfo failed")
		return common.AddressInfo{}, ErrRPCError
	}

	publicAddress := info.Address
	// Get confidential if request address is different
	if len(info.Confidential) > 0 && info.Confidential != info.Address {
		publicAddress = info.Confidential
	}

	result := common.AddressInfo{
		PublicAddress:  publicAddress,
		Unconfidential: info.Unconfidential,
	}

	log.WithFields(logrus.Fields{
		"PublicAddress":  result.PublicAddress,
		"Unconfidential": result.Unconfidential,
	}).Debug("Bitcoin RPC")

	return result, nil
}

func (p *BitcoinClient) ListUnspent(ctx context.Context, minConf, maxConf int, addresses ...string) ([]common.TransactionInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.ListUnspent")

	client := p.client
	if p.client == nil {
		return nil, ErrInternalError
	}

	var filter []commands.Address
	for _, addr := range addresses {
		filter = append(filter, commands.Address(addr))
	}

	if minConf > maxConf {
		minConf, maxConf = maxConf, minConf
	}

	list, err := commands.ListUnspentMinMaxAddresses(ctx, client.Client, minConf, maxConf, filter)
	if err != nil {
		log.WithError(err).
			Error("ListUnspentMinMaxAddresses failed")
		return nil, ErrRPCError
	}

	var result []common.TransactionInfo
	for _, tx := range list {
		result = append(result, common.TransactionInfo{
			Account:       tx.Label,
			Address:       string(tx.Address),
			Asset:         string(tx.Asset),
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
