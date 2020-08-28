// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/condensat/bank-core/wallet/ssm/commands"
	"github.com/ybbus/jsonrpc"
)

const chain = "bitcoin-test"
const entropy = "9e8ff5dd31e53502cfcd06e568cbdc881d63e598f8d28e90d403b4c986cf1e60"
const fingerprint = "548041a6"
const hdPathPrefix = "84h/0h"

type NewMasterResponse struct {
	Chain       string `json:"chain"`
	Fingerprint string `json:"fingerprint"`
}

type GetXpubResponse struct {
	Chain string `json:"chain"`
	Xpub  string `json:"xpub"`
}

type NewAddressResponse struct {
	Address string `json:"address"`
	Chain   string `json:"chain"`
	PubKey  string `json:"pubkey"`
}

type SignTxResponse struct {
	Chain    string `json:"chain"`
	SignedTx string `json:"signed_tx"`
}

func main() {
	ctx := context.Background()

	proxyURL, err := url.Parse("socks5://127.0.0.1:9050")
	if err != nil {
		panic(err)
	}
	const endpoint = "http://f3ughmonacr57liewfw6uzuzwu5vs3rl5znotyd525vcr2232gstbsid.onion/api/v1"
	rpcClient := jsonrpc.NewClientWithOpts(endpoint, &jsonrpc.RPCClientOpts{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
	})

	var actions []int
	batchSize := 32
	var count int
	for count < 1 {
		count++
		actions = append(actions, count)
	}

	batches := make([][]int, 0, (len(actions)+batchSize-1)/batchSize)

	for batchSize < len(actions) {
		actions, batches = actions[batchSize:], append(batches, actions[0:batchSize:batchSize])
	}
	batches = append(batches, actions)

	fmt.Printf("%d batches\n", len(batches))

	spendPath := ""

	for _, batch := range batches {
		var wg sync.WaitGroup
		for _, b := range batch {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				address, path, err := newAddress(ctx, rpcClient, chain, fingerprint, hdPathPrefix, id, false, 10)
				if err != nil {
					panic(err)
				}
				spendPath = path
				fmt.Printf("deposit address (%s): %+v\n", path, address)
				address, path, err = newAddress(ctx, rpcClient, chain, fingerprint, hdPathPrefix, id, true, 10)
				if err != nil {
					panic(err)
				}
				fmt.Printf("change address (%s): %+v\n", path, address)
			}(b)
		}
		wg.Wait()
	}

	// chain: str, tx: str, fingerprints: str, paths: str, values
	txToSign := "0200000001bf4ec7c318c2e61c33d02c2fc6fa9f6d56bf65aa07d46485a199c4194b0a53c60100000000feffffff021b610000000000001600142c6cb5f4191cde8bb1aa0b0d3683bb361075cba050c3000000000000160014bdef03be9f98c2af20b5caeaea617016da675c9400000000"
	var inputs []commands.SignTxInputs = []commands.SignTxInputs{{
		SsmPath: commands.SsmPath{
			Fingerprint: fingerprint,
			Path:        spendPath,
		},
		Amount: 0.0005},
	}
	signed, err := commands.SignTx(ctx, rpcClient, chain, txToSign, inputs...)
	if err != nil {
		panic(err)
	}
	fmt.Printf("sign_tx: %+v\n", signed)
}

func newAddress(ctx context.Context, rpcClient jsonrpc.RPCClient, chain, fingerprint, prefix string, id int, change bool, retry int) (commands.NewAddressResponse, string, error) {
	hdPath := fmt.Sprintf("%s/%d", prefix, id)
	if change {
		hdPath = fmt.Sprintf("%s/1", hdPath)
	}

	for retry > 0 {
		retry--

		address, err := commands.NewAddress(ctx, rpcClient, chain, fingerprint, hdPath)

		if err != nil {
			<-time.After(100 * time.Millisecond)
			continue
		}
		return address, hdPath, nil
	}

	return commands.NewAddressResponse{}, "", errors.New("RPC: newAddress failed")
}
