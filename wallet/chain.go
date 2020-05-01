// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"context"
	"errors"
	"sort"

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

type TransactionInfo struct {
	TxID          string
	Amount        float64
	Confirmations int64
}
type AddressInfo struct {
	Chain         string
	PublicAddress string
	Mined         uint64 // 0 unknown, 1 mempool, BlockHeight
	Transactions  []TransactionInfo
}

func GetNewAddress(ctx context.Context, chain, account string) (string, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.GetNewAddress")

	log = log.WithField("Chain", chain)

	client := ChainClientFromContext(ctx, chain)
	if client == nil {
		return "", ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return "", cache.ErrLockError
	}
	defer lock.Unlock()

	return client.GetNewAddress(ctx, account)
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

func FetchChainAddressesInfo(ctx context.Context, chain string, currentHeight, minConf, maxConf uint64, publicAddresses ...string) ([]AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.FetchChainAddresses")

	log = log.WithField("Chain", chain)

	client := ChainClientFromContext(ctx, chain)
	if client == nil {
		return nil, ErrChainClientNotFound
	}

	if len(publicAddresses) == 0 {
		log.Debug("No addresses provided")
		return nil, nil
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return nil, cache.ErrLockError
	}
	defer lock.Unlock()

	if minConf > maxConf {
		maxConf, minConf = minConf, maxConf
	}

	firsts := make(map[string]*AddressInfo)
	batches := batchAddresses(16, publicAddresses...)
	for _, batch := range batches {

		list, err := client.ListUnspent(ctx, int(minConf), int(maxConf), batch...)
		if err != nil {
			log.WithError(err).
				Error("Failed to ListUnspent")
			return nil, err
		}
		// Order oldest first
		sort.Slice(list, func(i, j int) bool {
			return list[i].Confirmations > list[j].Confirmations
		})

		for _, utxo := range list {
			// create if address is already not exists
			if _, ok := firsts[utxo.Address]; !ok {

				// zero confirmation mean in mempool
				var blockHeight uint64
				if utxo.Confirmations > 0 {
					blockHeight = currentHeight - uint64(utxo.Confirmations)
				}

				// create new map entry
				firsts[utxo.Address] = &AddressInfo{
					Chain:         chain,
					PublicAddress: utxo.Address,
					Mined:         blockHeight,
				}
			}

			// append TxID to transactions
			addr := firsts[utxo.Address]
			addr.Transactions = append(addr.Transactions, TransactionInfo{
				TxID:          utxo.TxID,
				Amount:        utxo.Amount,
				Confirmations: utxo.Confirmations,
			})
		}
	}

	var result []AddressInfo
	for _, utxo := range firsts {
		if utxo == nil {
			continue
		}
		result = append(result, *utxo)
	}

	return result, nil
}

func batchAddresses(batchSize int, addresses ...string) [][]string {
	if batchSize < 1 {
		batchSize = 1
	}
	if batchSize > 32 {
		batchSize = 32
	}
	batches := make([][]string, 0, (len(addresses)+batchSize-1)/batchSize)

	for batchSize < len(addresses) {
		addresses, batches = addresses[batchSize:], append(batches, addresses[0:batchSize:batchSize])
	}
	batches = append(batches, addresses)

	return batches[:]
}
