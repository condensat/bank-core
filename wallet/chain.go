// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"context"
	"errors"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/sirupsen/logrus"
)

var (
	ErrChainClientNotFound = errors.New("ChainClient Not Found")
)

type ChainState struct {
	Chain  string
	Height uint64
}

type AddressInfo struct {
	Chain         string
	PublicAddress string
	Mined         uint64 // 0 unknown, 1 mempool, BlockHeight
}

func FetchChainsState(ctx context.Context, chains ...string) ([]ChainState, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.FetchChainsState")

	var result []ChainState
	for _, chain := range chains {
		state, err := fetchChainState(ctx, chain)
		if err != nil {
			continue
		}

		result = append(result, state)
	}

	log.
		WithField("Count", len(result)).
		Debug("Chains State Fetched")

	return result, nil
}

func fetchChainState(ctx context.Context, chain string) (ChainState, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.fetchChainState")

	log = log.WithField("Chain", chain)

	client := ChainClientFromContext(ctx, chain)
	if client == nil {
		return ChainState{}, ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return ChainState{}, cache.ErrLockError
	}
	defer lock.Unlock()

	blockCount, err := client.GetBlockCount(ctx)
	if err != nil {
		return ChainState{}, err
	}

	log.
		WithFields(logrus.Fields{
			"Chain":  chain,
			"Height": blockCount,
		}).Info("Client RPC")

	return ChainState{
		Chain:  chain,
		Height: uint64(blockCount),
	}, nil
}

func FetchChainAddressesInfo(ctx context.Context, chain string, publicAddresses ...string) ([]AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.FetchChainAddresses")

	log = log.WithField("Chain", chain)

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return nil, cache.ErrLockError
	}
	defer lock.Unlock()

	var result []AddressInfo
	for _, publicAddress := range publicAddresses {

		// Todo: RPC call to chain daemon

		result = append(result, AddressInfo{
			Chain:         chain,
			PublicAddress: publicAddress,
			Mined:         42,
		})
	}

	return result, nil
}
