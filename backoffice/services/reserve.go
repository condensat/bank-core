// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"sort"

	"github.com/condensat/bank-core/utils"

	wallet "github.com/condensat/bank-core/wallet/client"
)

type WalletInfo struct {
	Chain  string  `json:"chain"`
	UTXOs  int     `json:"utxos"`
	Amount float64 `json:"amount"`
}

type WalletUTXO struct {
	TxID   string  `json:"txid"`
	Vout   int     `json:"vout"`
	Asset  string  `json:"asset"`
	Amount float64 `json:"amount"`
	Locked bool    `json:"locked"`
}

type WalletDetail struct {
	Chain  string       `json:"chain"`
	Height int          `json:"height"`
	UTXOs  []WalletUTXO `json:"utxos"`
}

type WalletStatus struct {
	Chain  string     `json:"chain"`
	Asset  string     `json:"asset"`
	Total  WalletInfo `json:"total"`
	Locked WalletInfo `json:"locked"`
}

type ReserveStatus struct {
	Wallets []WalletStatus `json:"wallets"`
}

func FetchReserveStatus(ctx context.Context) (ReserveStatus, error) {
	walletStatus, err := wallet.WalletStatus(ctx, wallet.WalletStatusWildcard)
	if err != nil {
		return ReserveStatus{}, err
	}

	var wallets []WalletStatus
	assetMap := make(map[string]*WalletStatus)
	for _, wallet := range walletStatus.Wallets {
		for _, utxo := range wallet.UTXOs {

			// get or create WalletStatus from assetMap
			key := wallet.Chain + utxo.Asset
			ws, ok := assetMap[key]
			if !ok {
				ws = &WalletStatus{
					Chain: wallet.Chain,
					Asset: utxo.Asset,
				}
				assetMap[key] = ws
			}

			ws.Total.Amount += utxo.Amount
			ws.Total.UTXOs++
			if utxo.Locked {
				ws.Locked.Amount += utxo.Amount
				ws.Locked.UTXOs++
			}
		}
	}

	for _, ws := range assetMap {
		ws.Total.Amount = utils.ToFixed(ws.Total.Amount, 8)
		ws.Locked.Amount = utils.ToFixed(ws.Locked.Amount, 8)

		wallets = append(wallets, *ws)
	}

	// Sort wallets
	sort.Slice(wallets, func(i, j int) bool {
		if wallets[i].Chain != wallets[j].Chain {
			return wallets[i].Chain < wallets[j].Chain
		}

		return wallets[i].Asset < wallets[j].Asset
	})

	return ReserveStatus{
		Wallets: wallets,
	}, nil
}

func FetchWalletList(ctx context.Context) ([]string, error) {
	return wallet.WalletList(ctx)
}

func FetchWalletDetail(ctx context.Context, chain string) ([]WalletDetail, error) {
	status, err := wallet.WalletStatus(ctx, chain)
	if err != nil {
		return nil, err
	}

	var result []WalletDetail
	for _, wallet := range status.Wallets {
		// skip non requested wallet
		if wallet.Chain != chain {
			continue
		}

		var utxos []WalletUTXO
		for _, utxo := range wallet.UTXOs {
			utxos = append(utxos, WalletUTXO{
				TxID:   utxo.TxID,
				Vout:   utxo.Vout,
				Asset:  utxo.Asset,
				Amount: utxo.Amount,
				Locked: utxo.Locked,
			})
		}
		result = append(result, WalletDetail{
			Chain:  wallet.Chain,
			Height: wallet.Height,
			UTXOs:  utxos,
		})
	}

	return result, nil
}
