// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package chain

import (
	"context"
	"errors"
	"math/rand"
	"sort"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/ssm/commands"

	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

const (
	AddressBatchSize = 16 // maximum address count for RPC requests

	MaxConf = 9999999
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
	Vout          int64
	Asset         string
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

	client := common.ChainClientFromContext(ctx, chain)
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

func ImportAddress(ctx context.Context, chain, account, address, pubkey, blindingkey string) error {
	log := logger.Logger(ctx).WithField("Method", "wallet.ImportAddress")

	log = log.WithField("Chain", chain)

	client := common.ChainClientFromContext(ctx, chain)
	if client == nil {
		return ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return cache.ErrLockError
	}
	defer lock.Unlock()

	return client.ImportAddress(ctx, account, address, pubkey, blindingkey)
}

func GetAddressInfo(ctx context.Context, chain, address string) (common.AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.GetAddressInfo")

	log = log.WithField("Chain", chain)

	client := common.ChainClientFromContext(ctx, chain)
	if client == nil {
		return common.AddressInfo{}, ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return common.AddressInfo{}, cache.ErrLockError
	}
	defer lock.Unlock()

	info, err := client.GetAddressInfo(ctx, address)
	if err != nil {
		log.WithError(err).
			Error("Failed to GetAddressInfo")
		return common.AddressInfo{}, cache.ErrLockError
	}

	info.Chain = chain

	return info, nil
}

func WalletInfo(ctx context.Context, chain string) (common.WalletInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.WalletInfo")

	log = log.WithField("Chain", chain)

	client := common.ChainClientFromContext(ctx, chain)
	if client == nil {
		return common.WalletInfo{}, ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return common.WalletInfo{}, cache.ErrLockError
	}
	defer lock.Unlock()

	unspent, err := client.ListUnspent(ctx, 0, MaxConf)
	if err != nil {
		log.WithError(err).
			Error("Failed to ListUnspent")
		return common.WalletInfo{}, err
	}

	var utxos []common.UTXOInfo

	// Available utxos
	for _, utxo := range unspent {
		utxos = append(utxos, common.UTXOInfo{
			TxID:   utxo.TxID,
			Vout:   int(utxo.Vout),
			Asset:  utxo.Asset,
			Amount: utxo.Amount,
			Locked: false,
		})
	}

	// Locked utxos
	locked, err := client.ListLockUnspent(ctx)
	if err != nil {
		log.WithError(err).
			Error("Failed to ListLockUnspent")
		return common.WalletInfo{}, err
	}
	for _, utxo := range locked {
		utxos = append(utxos, common.UTXOInfo{
			TxID:   utxo.TxID,
			Vout:   int(utxo.Vout),
			Amount: utxo.Amount,
			Locked: true,
		})
	}

	// create & return result
	return common.WalletInfo{
		Chain: chain,
		UTXOs: utxos,
	}, nil
}

func LockUnspent(ctx context.Context, chain string, unlock bool, utxos ...common.TransactionInfo) error {
	log := logger.Logger(ctx).WithField("Method", "wallet.LockUnspent")

	log = log.WithField("Chain", chain)

	client := common.ChainClientFromContext(ctx, chain)
	if client == nil {
		return ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return cache.ErrLockError
	}
	defer lock.Unlock()

	err = client.LockUnspent(ctx, unlock, utxos...)
	if err != nil {
		log.WithError(err).
			Error("Failed to LockUnspent")
		return cache.ErrLockError
	}

	return nil
}

func ListLockUnspent(ctx context.Context, chain string) ([]common.TransactionInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.ListLockUnspent")

	log = log.WithField("Chain", chain)

	client := common.ChainClientFromContext(ctx, chain)
	if client == nil {
		return nil, ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return nil, cache.ErrLockError
	}
	defer lock.Unlock()

	list, err := client.ListLockUnspent(ctx)
	if err != nil {
		log.WithError(err).
			Error("Failed to ListLockUnspent")
		return nil, err
	}

	return list, nil
}

func GetTransaction(ctx context.Context, chain string, txID string) (common.TransactionInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.GetTransaction")

	log = log.WithField("Chain", chain)

	client := common.ChainClientFromContext(ctx, chain)
	if client == nil {
		return common.TransactionInfo{}, ErrChainClientNotFound
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return common.TransactionInfo{}, cache.ErrLockError
	}
	defer lock.Unlock()

	result, err := client.GetTransaction(ctx, txID)
	if err != nil {
		log.WithError(err).
			Error("Failed to GetTransaction")
		return common.TransactionInfo{}, cache.ErrLockError
	}

	return result, nil
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

	client := common.ChainClientFromContext(ctx, chain)
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

func FetchChainAddressesInfo(ctx context.Context, state ChainState, minConf, maxConf uint64, publicAddresses ...string) ([]AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.FetchChainAddresses")

	log = log.WithFields(logrus.Fields{
		"Chain":  state.Chain,
		"Height": state.Height,
	})

	client := common.ChainClientFromContext(ctx, state.Chain)
	if client == nil {
		return nil, ErrChainClientNotFound
	}

	if len(publicAddresses) == 0 {
		log.Debug("No addresses provided")
		return nil, nil
	}

	// Acquire Lock
	lock, err := cache.LockChain(ctx, state.Chain)
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
	batches := utils.BatchString(AddressBatchSize, publicAddresses...)
	for _, batch := range batches {

		// RPC requets
		list, err := client.ListUnspent(ctx, int(minConf), int(maxConf), batch...)
		if err != nil {
			log.WithError(err).
				Error("Failed to ListUnspent")
			return nil, err
		}

		lockedUtxos, err := client.ListLockUnspent(ctx)
		if err != nil {
			log.WithError(err).
				Error("Failed to ListLockUnspent")
			return nil, err
		}
		log.WithField("LockedUtxos", lockedUtxos).
			Trace("ListLockUnspent")

		for _, utxo := range lockedUtxos {
			tx, err := client.GetTransaction(ctx, utxo.TxID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetTransaction")
				return nil, err
			}
			log.WithField("Tx", tx).
				Trace("Locked Tx")

			list = append(list, tx)
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
					blockHeight = state.Height - uint64(utxo.Confirmations)
				}

				// create new map entry
				firsts[utxo.Address] = &AddressInfo{
					Chain:         state.Chain,
					PublicAddress: utxo.Address,
					Mined:         blockHeight,
				}
			}

			// append TxID to transactions
			addr := firsts[utxo.Address]
			addr.Transactions = append(addr.Transactions, TransactionInfo{
				TxID:          utxo.TxID,
				Vout:          utxo.Vout,
				Asset:         utxo.Asset,
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

func SpendFunds(ctx context.Context, chain string, changeAddress string, spendInfos []common.SpendInfo) (common.SpendTx, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.SpendFunds")

	log = log.WithField("Chain", chain)

	client := common.ChainClientFromContext(ctx, chain)
	if client == nil {
		return common.SpendTx{}, ErrChainClientNotFound
	}

	// Create, Fund, Sign & Broadcast transaction
	blindTransaction := blindTransactionFromChain(chain)
	var inputs []common.UTXOInfo
	if blindTransaction {

		// Sum up amounts for each assets
		assetTotalSpend := make(map[string]float64)
		for _, spendInfo := range spendInfos {
			assetHash := spendInfo.Asset.Hash
			if len(assetHash) == 0 {
				continue
			}
			if _, exists := assetTotalSpend[assetHash]; !exists {
				assetTotalSpend[assetHash] = 0.0
			}

			assetTotalSpend[assetHash] = utils.ToFixed(assetTotalSpend[assetHash]+spendInfo.Amount, 8)
		}

		// Compute change amount for asset
		// only one spendInfo contain a ChangeAddress for each asset
		for i, spendInfo := range spendInfos {
			changeAddress := spendInfo.Asset.ChangeAddress
			if len(changeAddress) == 0 {
				continue
			}

			assetHash := spendInfo.Asset.Hash
			transactions, err := client.ListUnspentByAsset(ctx, 0, 999999, assetHash)
			if err != nil {
				log.WithError(err).
					WithField("Asset", assetHash).
					Error("Failed to ListUnspentByAsset")
				return common.SpendTx{}, err
			}

			log.WithFields(logrus.Fields{
				"AssetHash":    assetHash,
				"Transactions": transactions,
			}).Trace("UTXOs")

			// shuffle UTXO to spent
			rand.Shuffle(len(transactions), func(i, j int) {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			})

			totalSpendAmount := assetTotalSpend[assetHash]
			var totalAmount float64
			for _, transaction := range transactions {
				if transaction.Asset != assetHash {
					log.WithError(err).
						WithFields(logrus.Fields{
							"Asset":            assetHash,
							"TransactionAsset": transaction.Asset,
						}).Error("Asset Hash missmatch")

					return common.SpendTx{}, err
				}
				totalAmount = utils.ToFixed(totalAmount+transaction.Amount, 8)

				inputs = append(inputs, common.UTXOInfo{
					TxID: transaction.TxID,
					Vout: int(transaction.Vout),
				})

				if totalAmount >= totalSpendAmount {
					// update changeAmount
					spendInfo.Asset.ChangeAmount = utils.ToFixed(totalAmount-totalSpendAmount, 8)
					spendInfos[i] = spendInfo
					log.WithFields(logrus.Fields{
						"SpendInfo":         spendInfo,
						"TotalSpendAmount":  totalSpendAmount,
						"ChangeAmount":      spendInfo.Asset.ChangeAmount,
						"TransactionAmount": transaction.Amount,
						"TotalAmount":       totalAmount,
					}).Trace("ChangeAmount")
					break
				}
			}

			if totalAmount <= 0.0 {
				log.WithFields(logrus.Fields{
					"SpendInfo":   spendInfo,
					"TotalAmount": totalAmount,
				}).Warning("No change created")
			}
		}
	}
	log.WithFields(logrus.Fields{
		"SpendInfos": spendInfos,
		"Inputs":     inputs,
	}).Trace("Spend Infos")

	tx, err := client.SpendFunds(ctx, changeAddress, inputs, spendInfos, getAddressInfoFromDatabase, blindTransaction)
	if err != nil {
		log.WithError(err).
			Error("Failed to SpendFunds")
		return common.SpendTx{}, err
	}

	return tx, nil
}

func getAddressInfoFromDatabase(ctx context.Context, address string, isUnconfidential bool) (commands.SsmPath, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.chain.getAddressInfoFromDatabase")
	db := appcontext.Database(ctx)

	if len(address) == 0 {
		return commands.SsmPath{}, errors.New("Invalid address")
	}

	if isUnconfidential {
		cryptoAddress, err := database.GetCryptoAddressWithUnconfidential(db, model.String(address))
		if err != nil {
			log.WithError(err).
				Error("Failed to GetCryptoAddressWithUnconfidential")
			return commands.SsmPath{}, err
		}

		// get public address for ssm database request
		address = string(cryptoAddress.PublicAddress)
	}

	ssmAddress, err := database.GetSsmAddressByPublicAddress(db, model.SsmPublicAddress(address))
	if err != nil {
		log.WithError(err).
			Error("Failed to GetSsmAddressByPublicAddress")
		return commands.SsmPath{}, err
	}
	addressInfo, err := database.GetSsmAddressInfo(db, ssmAddress.ID)
	if err != nil {
		log.WithError(err).
			Error("Failed to GetSsmAddressInfo")
		return commands.SsmPath{}, err
	}

	return commands.SsmPath{
		Chain:       string(addressInfo.Chain),
		Fingerprint: string(addressInfo.Fingerprint),
		Path:        string(addressInfo.HDPath),
	}, nil
}

func blindTransactionFromChain(chain string) bool {
	switch chain {
	case "liquid-mainnet":
		return true

	default:
		return false
	}
}
