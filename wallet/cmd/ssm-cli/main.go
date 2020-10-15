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
	"os"
	"sync"
	"time"

	"github.com/ybbus/jsonrpc"

	"github.com/condensat/bank-core/wallet/common"
	ssmCommands "github.com/condensat/bank-core/wallet/ssm/commands"

	btcCommands "github.com/condensat/bank-core/wallet/bitcoin/commands"
	btcrpc "github.com/condensat/bank-core/wallet/rpc"
)

const chain = "bitcoin-test"
const entropy = "9e8ff5dd31e53502cfcd06e568cbdc881d63e598f8d28e90d403b4c986cf1e60"
const fingerprint = "548041a6"
const hdPathPrefix = "84h/0h"
const label = "condensat"

type NewMasterresponse struct {
	Chain       string `json:"chain"`
	Fingerprint string `json:"fingerprint"`
}

type GetXpubResponse struct {
	Chain string `json:"chain"`
	Xpub  string `json:"xpub"`
}

func bitcoinClient() jsonrpc.RPCClient {
	return btcrpc.New(btcrpc.Options{
		ServerOptions: common.ServerOptions{Protocol: "http", HostName: "bitcoin-testnet", Port: 18332},
		User:          "condensat",
		Password:      "condensat",
	}).Client

}

func cryptoSsmClient() jsonrpc.RPCClient {
	// crypto-ssm is on tor
	proxyURL, err := url.Parse("socks5://127.0.0.1:9050")
	if err != nil {
		panic(err)
	}
	const endpoint = "http://f3ughmonacr57liewfw6uzuzwu5vs3rl5znotyd525vcr2232gstbsid.onion/api/v1"
	return jsonrpc.NewClientWithOpts(endpoint, &jsonrpc.RPCClientOpts{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
	})

}

func main() {
	ctx := context.Background()
	btcClient := bitcoinClient()
	ssmClient := cryptoSsmClient()

	if len(os.Args) > 1 {
		GenerateAddresses(ctx, ssmClient, btcClient)
		return
	}

	// mock database
	addressPaths := map[string]string{
		"tb1q93kttaqern0ghvd2pvxndqamxcg8tjaq4hq6xu": "84h/0h/1",
		"tb1q29neqwhha0a94de7j7vewz4hkzwup5c75jz99u": "84h/0h/1/1",
	}

	// Create & Fund Transaction
	txToSign := func() string {
		hex, err := btcCommands.CreateRawTransaction(ctx, btcClient, nil, []btcCommands.SpendInfo{
			{Address: "tb1qhhhs805lnrp27g94et4w5ctszmdxwhy5e90gw9", Amount: 0.001},
		}, nil)
		if err != nil {
			panic(err)
		}

		funded, err := btcCommands.FundRawTransactionWithOptions(ctx, btcClient, hex, btcCommands.FundRawTransactionOptions{
			ChangeAddress:          "tb1q29neqwhha0a94de7j7vewz4hkzwup5c75jz99u",
			ChangePosition:         0,
			SubtractFeeFromOutputs: []int{0},
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("FundRawTransaction: %s\n", funded.Hex)

		return funded.Hex
	}()

	rawTx, err := btcCommands.DecodeRawTransaction(ctx, btcClient, btcCommands.Transaction(txToSign))
	if err != nil {
		panic(err)
	}

	transaction, err := btcCommands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		panic(err)
	}

	// grab inputs path & amouts
	var inputs []ssmCommands.SignTxInputs
	for _, in := range transaction.Vin {
		txID := btcCommands.TransactionID(in.Txid)
		txHex, err := btcCommands.GetRawTransaction(ctx, btcClient, txID)
		if err != nil {
			panic(err)
		}
		rawTxIn, err := btcCommands.DecodeRawTransaction(ctx, btcClient, btcCommands.Transaction(txHex))
		if err != nil {
			panic(err)
		}
		tx, err := btcCommands.ConvertToRawTransactionBitcoin(rawTxIn)
		if err != nil {
			panic(err)
		}

		// append input entry
		out := tx.Vout[in.Vout]
		amount := out.Value
		address := addressPaths[out.ScriptPubKey.Addresses[0]]
		inputs = append(inputs, ssmCommands.SignTxInputs{
			SsmPath: ssmCommands.SsmPath{
				Chain:       "bitcoin",
				Fingerprint: fingerprint,
				Path:        address,
			},
			Amount: amount,
		})
	}

	// Sign Transaction
	signedTx := func() string {
		signed, err := ssmCommands.SignTx(ctx, ssmClient, chain, txToSign, inputs...)
		if err != nil {
			panic(err)
		}
		fmt.Printf("sign_tx: %+v\n", signed)

		return signed.SignedTx
	}()

	// Broadcast Transaction
	txID, err := btcCommands.SendRawTransaction(ctx, btcClient, btcCommands.Transaction(signedTx))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction sent: %s\n", txID)
}

func GenerateAddresses(ctx context.Context, ssmClient, btcClient jsonrpc.RPCClient) {
	var hdMaster NewMasterresponse
	err := ssmClient.CallFor(&hdMaster, "new_master",
		chain,
		entropy,
		true,
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("new_master: %+v\n", hdMaster)

	var xpub GetXpubResponse
	err = ssmClient.CallFor(&xpub, "get_xpub",
		chain,
		hdMaster.Fingerprint,
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get_xpub: %+v\n", xpub)

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

	for _, batch := range batches {
		var wg sync.WaitGroup
		for _, b := range batch {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				address, path, err := newAddress(ctx, ssmClient, hdMaster.Chain, hdMaster.Fingerprint, hdPathPrefix, id, false, 10)
				if err != nil {
					panic(err)
				}
				err = ImportAddress(ctx, btcClient, address, label)
				if err != nil {
					panic(err)
				}

				fmt.Printf("deposit address (%s): %+v\n", path, address)
				address, path, err = newAddress(ctx, ssmClient, hdMaster.Chain, hdMaster.Fingerprint, hdPathPrefix, id, true, 10)
				if err != nil {
					panic(err)
				}
				err = ImportAddress(ctx, btcClient, address, label)
				if err != nil {
					panic(err)
				}
				fmt.Printf("change address (%s): %+v\n", path, address)
			}(b)
		}
		wg.Wait()
	}
}

func ImportAddress(ctx context.Context, btcClient jsonrpc.RPCClient, address ssmCommands.NewAddressResponse, label string) error {
	err := btcCommands.ImportAddress(ctx, btcClient, btcCommands.Address(address.Address), label, false)
	if err != nil {
		return err
	}
	err = btcCommands.ImportPubKey(ctx, btcClient, btcCommands.PubKey(address.PubKey), label, false)
	if err != nil {
		return err
	}

	return nil
}

func newAddress(ctx context.Context, ssmClient jsonrpc.RPCClient, chain, fingerprint, prefix string, id int, change bool, retry int) (ssmCommands.NewAddressResponse, string, error) {
	hdPath := fmt.Sprintf("%s/%d", prefix, id)
	if change {
		hdPath = fmt.Sprintf("%s/1", hdPath)
	}

	for retry > 0 {
		retry--

		address, err := ssmCommands.NewAddress(ctx, ssmClient, chain, fingerprint, hdPath)

		if err != nil {
			<-time.After(100 * time.Millisecond)
			continue
		}
		return address, hdPath, nil
	}

	return ssmCommands.NewAddressResponse{}, "", errors.New("RPC: newAddress failed")
}
