// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"log"
	"os"

	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/rpc"

	"github.com/condensat/bank-core/wallet/bitcoin/commands"

	dotenv "github.com/joho/godotenv"
)

func init() {
	_ = dotenv.Load()
}

func main() {
	ctx := context.Background()
	RawTransaction(ctx)
}

func RawTransaction(ctx context.Context) {
	rpcClient := bitcoinRpcClient("bitcoin-testnet", 18332)
	if rpcClient == nil {
		panic("Invalid rpcClient")
	}

	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: "tb1qqjv0dec9vagycgwpchdkxsnapl9uy92dek4nau", Amount: 0.000003},
	}, nil)
	if err != nil {
		panic(err)
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	rawTx, err := commands.DecodeRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		panic(err)
	}
	decoded, err := commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		panic(err)
	}
	log.Printf("DecodeRawTransaction: %+v\n", decoded)

	funded, err := commands.FundRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		panic(err)
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	rawTx, err = commands.DecodeRawTransaction(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		panic(err)
	}
	decoded, err = commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		panic(err)
	}
	log.Printf("FundRawTransaction Hex: %+v\n", decoded)

	addressMap := make(map[commands.Address]commands.Address)
	for _, in := range decoded.Vin {

		txInfo, err := commands.GetTransaction(ctx, rpcClient, in.Txid, true)
		if err != nil {
			panic(err)
		}

		addressMap[txInfo.Address] = txInfo.Address
		for _, d := range txInfo.Details {
			address := commands.Address(d.Address)
			addressMap[address] = address
		}
	}

	signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		panic(err)
	}
	if !signed.Complete {
		panic("SignRawTransactionWithWallet failed")
	}
	log.Printf("SignRawTransactionWithWallet: %+v\n", signed.Hex)

	txId, err := commands.SendRawTransaction(ctx, rpcClient, commands.Transaction(signed.Hex))
	if err != nil {
		panic(err)
	}
	log.Printf("SendRawTransaction: %+v\n", txId)
}

func bitcoinRpcClient(hostname string, port int) commands.RpcClient {
	password := os.Getenv("BITCOIN_TESTNET_PASSWORD")
	return rpc.New(rpc.Options{
		ServerOptions: common.ServerOptions{Protocol: "http", HostName: hostname, Port: port},
		User:          "bank-wallet",
		Password:      password,
	}).Client
}
