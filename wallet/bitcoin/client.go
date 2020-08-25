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
	ssmCommands "github.com/condensat/bank-core/wallet/ssm/commands"

	"github.com/sirupsen/logrus"
	"github.com/ybbus/jsonrpc"
)

var (
	ErrInternalError    = errors.New("Internal Error")
	ErrRPCError         = errors.New("RPC Error")
	ErrInvalidAccount   = errors.New("Invalid Account")
	ErrInvalidAddress   = errors.New("Invalid Address format")
	ErrInvalidPubKey    = errors.New("Invalid PubKey")
	ErrLockUnspentFails = errors.New("LockUnpent Failed")
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

func (p *BitcoinClient) ImportAddress(ctx context.Context, account, address, pubkey, blindingkey string) error {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.InmportAddress")

	client := p.client
	if p.client == nil {
		return ErrInternalError
	}
	if len(address) == 0 {
		return ErrInvalidAddress
	}
	if len(pubkey) == 0 {
		return ErrInvalidPubKey
	}

	err := commands.ImportAddress(ctx, client.Client, commands.Address(address), account, false)
	if err != nil {
		log.WithError(err).
			Error("ImportAddress failed")
		return ErrRPCError
	}

	err = commands.ImportPubKey(ctx, client.Client, commands.PubKey(pubkey), account, false)
	if err != nil {
		log.WithError(err).
			Error("ImportPubKey failed")
		return ErrRPCError
	}

	// optional blindingkey for liquid clients
	if len(blindingkey) > 0 {
		err = commands.ImportBlindingKey(ctx, client.Client, commands.Address(address), commands.BlindingKey(blindingkey))
		if err != nil {
			log.WithError(err).
				Error("ImportPubKey failed")
			return ErrRPCError
		}
	}

	log.
		WithFields(logrus.Fields{
			"PubKey":      pubkey,
			"Address":     address,
			"BlindingKey": blindingkey,
		}).Debug("Bitcoin RPC")

	return nil
}

func (p *BitcoinClient) GetAddressInfo(ctx context.Context, address string) (common.AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.GetAddressInfo")

	client := p.client
	if p.client == nil {
		return common.AddressInfo{}, ErrInternalError
	}
	if len(address) == 0 {
		return common.AddressInfo{}, ErrInvalidAddress
	}

	log = log.WithField("Address", address)

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
		IsValid:        len(info.ScriptPubKey) != 0,
	}

	log.WithFields(logrus.Fields{
		"PublicAddress":  result.PublicAddress,
		"Unconfidential": result.Unconfidential,
	}).Debug("Bitcoin RPC")

	return result, nil
}

func (p *BitcoinClient) ListUnspent(ctx context.Context, minConf, maxConf int, addresses ...string) ([]common.TransactionInfo, error) {
	var filter []commands.Address
	for _, addr := range addresses {
		filter = append(filter, commands.Address(addr))
	}

	if minConf > maxConf {
		minConf, maxConf = maxConf, minConf
	}

	return p.listUnspent(ctx, minConf, maxConf, filter, "")
}

func (p *BitcoinClient) ListUnspentByAsset(ctx context.Context, minConf, maxConf int, asset string) ([]common.TransactionInfo, error) {
	var filter []commands.Address

	if minConf > maxConf {
		minConf, maxConf = maxConf, minConf
	}

	return p.listUnspent(ctx, minConf, maxConf, filter, asset)
}

func (p *BitcoinClient) listUnspent(ctx context.Context, minConf, maxConf int, filter []commands.Address, asset string) ([]common.TransactionInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.listUnspent")
	client := p.client
	if p.client == nil {
		return nil, ErrInternalError
	}

	list, err := commands.ListUnspentMinMaxAddressesAndOptions(ctx,
		client.Client,
		minConf, maxConf, filter,
		commands.ListUnspentOption{
			Asset: asset,
		},
	)
	if err != nil {
		log.WithError(err).
			Error("ListUnspentMinMaxAddressesAndOptions failed")
		return nil, ErrRPCError
	}

	var result []common.TransactionInfo
	for _, tx := range list {
		result = append(result, convertTransactionInfo(tx))
	}

	log.
		WithField("Count", len(list)).
		Debug("Bitcoin RPC")

	return result, nil
}

func (p *BitcoinClient) LockUnspent(ctx context.Context, unlock bool, transactions ...common.TransactionInfo) error {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.LockUnspent")

	client := p.client
	if p.client == nil {
		return ErrInternalError
	}

	var utxos []commands.UTXOInfo
	for _, tx := range transactions {
		utxos = append(utxos, commands.UTXOInfo{
			TxID: tx.TxID,
			Vout: int(tx.Vout),
		})
	}

	success, err := commands.LockUnspent(ctx, client.Client, unlock, utxos)
	if err != nil {
		log.WithError(err).
			Error("LockUnspent failed")
		return ErrRPCError
	}

	if !success {
		return ErrLockUnspentFails
	}

	return nil
}

func (p *BitcoinClient) ListLockUnspent(ctx context.Context) ([]common.TransactionInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.ListLockUnspent")

	client := p.client
	if p.client == nil {
		return nil, ErrInternalError
	}

	list, err := commands.ListLockUnspent(ctx, client.Client)
	if err != nil {
		log.WithError(err).
			Error("LockUnspent failed")
		return nil, ErrRPCError
	}

	var result []common.TransactionInfo
	for _, tx := range list {
		result = append(result, common.TransactionInfo{
			TxID: tx.TxID,
			Vout: int64(tx.Vout),
		})
	}

	return result, nil
}

func (p *BitcoinClient) GetTransaction(ctx context.Context, txID string) (common.TransactionInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.GetTransaction")

	client := p.client
	if p.client == nil {
		return common.TransactionInfo{}, ErrInternalError
	}

	tx, err := commands.GetTransaction(ctx, client.Client, txID, true)
	if err != nil {
		log.WithError(err).
			Error("GetTransaction failed")
		return common.TransactionInfo{}, ErrRPCError
	}

	return convertTransactionInfo(tx), nil
}

func (p *BitcoinClient) SpendFunds(ctx context.Context, changeAddress string, inputs []common.UTXOInfo, outputs []common.SpendInfo, addressInfo common.GetAddressInfo, blindTransaction bool) (common.SpendTx, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.SpendFunds")

	cryptoMode := common.CryptoModeFromContext(ctx)
	log = log.WithField("CryptoMode", cryptoMode)

	client := p.client
	if p.client == nil {
		return common.SpendTx{}, ErrInternalError
	}

	if inputs == nil {
		inputs = []common.UTXOInfo{}
	}

	in := convertUTXOInfo(inputs...)
	out, assets := convertSpendInfo(outputs...)

	log.WithFields(logrus.Fields{
		"inputs":  inputs,
		"outputs": outputs,
		"in":      in,
		"out":     out,
		"assets":  assets,
	}).Trace("convertSpendInfo")

	// Create transaction with no input
	hex, err := commands.CreateRawTransaction(ctx, client.Client, in, out, assets)
	if err != nil {
		log.WithError(err).
			Error("GetTransaction failed")
		return common.SpendTx{}, ErrRPCError
	}

	// Fund transaction (bitcoin-core will select inputs automatically)
	funded, err := fundRawTransactionWithCryptoMode(ctx, client, cryptoMode, changeAddress, hex)
	if err != nil {
		log.WithError(err).
			Error("FundRawTransaction failed")
		return common.SpendTx{}, ErrRPCError
	}

	// blind transaction if required
	txToSign := funded.Hex
	if blindTransaction {
		blinded, err := commands.BlindRawTransaction(ctx, client.Client, commands.Transaction(txToSign))
		if err != nil {
			log.WithError(err).
				Error("BlindRawTransaction failed")
			return common.SpendTx{}, ErrRPCError
		}

		txToSign = string(blinded)
	}

	// Sign transaction
	signed, err := signRawTransactionWithCryptoMode(ctx, client.Client, cryptoMode, txToSign, addressInfo, blindTransaction)
	if err != nil {
		log.WithError(err).
			WithField("TxToSign", txToSign).
			Error("signRawTransactionWithCryptoMode failed")
		return common.SpendTx{}, ErrRPCError
	}
	if !signed.Complete {
		log.Error("signRawTransactionWithCryptoMode not Complete")
		return common.SpendTx{}, ErrRPCError
	}

	log.WithError(err).
		WithFields(logrus.Fields{
			"Tx": signed.Hex,
		}).
		Trace("Transaction Signed")

	// Broadcast transaction to network
	tx, err := commands.SendRawTransaction(ctx, client.Client, commands.Transaction(signed.Hex))
	if err != nil {
		log.WithError(err).
			Error("SendRawTransaction failed")
		return common.SpendTx{}, ErrRPCError
	}

	// return TxID
	return common.SpendTx{
		TxID: string(tx),
	}, nil
}

func fundRawTransactionWithCryptoMode(ctx context.Context, client *rpc.Client, cryptoMode common.CryptoMode, changeAddress string, hex commands.Transaction) (commands.FundedTransaction, error) {
	switch cryptoMode {
	case common.CryptoModeCryptoSsm:
		if changeAddress == "" {
			return commands.FundedTransaction{}, errors.New("Invalid Change Address")
		}
		return commands.FundRawTransactionWithOptions(ctx,
			client.Client,
			hex,
			commands.FundRawTransactionOptions{
				ChangeAddress:   changeAddress,
				IncludeWatching: true,
			},
		)
	default:
		return commands.FundRawTransaction(ctx, client.Client, hex)
	}
}

func signRawTransactionWithCryptoMode(ctx context.Context, client jsonrpc.RPCClient, cryptoMode common.CryptoMode, txToSign string, addressInfo common.GetAddressInfo, blindedTransaction bool) (commands.SignedTransaction, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.signRawTransactionWithCryptoMode")

	switch cryptoMode {
	case common.CryptoModeCryptoSsm:
		const device = "crypto-ssm"
		ssmClient := common.SsmClientFromContext(ctx, device)
		if ssmClient == nil {
			return commands.SignedTransaction{}, errors.New("SSM not found")
		}

		if addressInfo == nil {
			return commands.SignedTransaction{}, errors.New("Invalid sign Callback")
		}
		rawTx, err := commands.DecodeRawTransaction(ctx, client, commands.Transaction(txToSign))
		if err != nil {
			return commands.SignedTransaction{}, errors.New("Failed to DecodeRawTransaction")
		}

		var inputs []ssmCommands.SignTxInputs
		chain := ""
		if !blindedTransaction {
			transaction, err := commands.ConvertToRawTransactionBitcoin(rawTx)
			if err != nil {
				return commands.SignedTransaction{}, errors.New("Failed to ConvertToRawTransactionBitcoin")
			}
			// grab inputs path & amouts
			for _, in := range transaction.Vin {
				txID := commands.TransactionID(in.Txid)
				txHex, err := commands.GetRawTransaction(ctx, client, txID)
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to GetRawTransaction")
				}
				rawTxIn, err := commands.DecodeRawTransaction(ctx, client, commands.Transaction(txHex))
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to DecodeRawTransaction")
				}
				tx, err := commands.ConvertToRawTransactionBitcoin(rawTxIn)
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to ConvertToRawTransactionBitcoin")
				}

				// append input entry
				out := tx.Vout[in.Vout]
				log.WithFields(logrus.Fields{
					"Txid": in.Txid,
					"Vout": in.Vout,
					"data": out,
				}).Trace("Spend UTXO")
				amount := out.Value
				address := out.ScriptPubKey.Addresses[0]
				info, err := addressInfo(ctx, address, false)
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to get address info")
				}

				// select chain from first input
				if len(chain) == 0 {
					chain = info.Chain
				} else if info.Chain != chain {
					return commands.SignedTransaction{}, errors.New("Chain missmatch")
				}

				inputs = append(inputs, ssmCommands.SignTxInputs{
					SsmPath: ssmCommands.SsmPath{
						Chain:       info.Chain,
						Fingerprint: info.Fingerprint,
						Path:        info.Path,
					},
					Amount: amount,
				})
			}
		} else {
			transaction, err := commands.ConvertToRawTransactionLiquid(rawTx)
			if err != nil {
				return commands.SignedTransaction{}, errors.New("Failed to ConvertToRawTransactionLiquid")
			}
			// grab inputs path & amouts
			for _, in := range transaction.Vin {
				txID := commands.TransactionID(in.Txid)
				txHex, err := commands.GetRawTransaction(ctx, client, txID)
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to GetRawTransaction")
				}
				rawTxIn, err := commands.DecodeRawTransaction(ctx, client, commands.Transaction(txHex))
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to DecodeRawTransaction")
				}
				tx, err := commands.ConvertToRawTransactionLiquid(rawTxIn)
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to ConvertToRawTransactionLiquid")
				}

				// append input entry
				out := tx.Vout[in.Vout]
				valueCommitment := out.Valuecommitment
				var value float64
				if len(valueCommitment) == 0 {
					value = out.Value
				}
				// at least one of value and valueCommitment musr be valid
				if len(valueCommitment) == 0 && value <= 0.0 {
					return commands.SignedTransaction{}, errors.New("Invalid value or valueCommitment")
				}

				address := out.ScriptPubKey.Addresses[0]
				info, err := addressInfo(ctx, address, true)
				if err != nil {
					return commands.SignedTransaction{}, errors.New("Failed to get address info")
				}

				// select chain from first input
				if len(chain) == 0 {
					chain = info.Chain
				} else if info.Chain != chain {
					return commands.SignedTransaction{}, errors.New("Chain missmatch")
				}

				inputs = append(inputs, ssmCommands.SignTxInputs{
					SsmPath: ssmCommands.SsmPath{
						Chain:       info.Chain,
						Fingerprint: info.Fingerprint,
						Path:        info.Path,
					},
					ValueCommitment: valueCommitment,
					Amount:          value, // value is valid when valueCommitment is empty
				})
			}
		}

		if len(chain) == 0 {
			return commands.SignedTransaction{}, errors.New("Invalid chain")
		}

		log.WithField("Inputs", inputs).
			Trace("Sign Input")

		// Sign Transaction
		signedTx, err := ssmClient.SignTx(ctx, chain, txToSign, inputs...)

		return commands.SignedTransaction{
			Complete: len(signedTx) > 0,
			Hex:      signedTx,
		}, err

	default:
		return commands.SignRawTransactionWithWallet(ctx, client, commands.Transaction(txToSign))
	}
}

func convertTransactionInfo(tx commands.TransactionInfo) common.TransactionInfo {
	return common.TransactionInfo{
		Account:       tx.Label,
		Address:       string(tx.Address),
		Asset:         string(tx.Asset),
		TxID:          tx.TxID,
		Vout:          int64(tx.Vout),
		Amount:        tx.Amount,
		Confirmations: tx.Confirmations,
		Spendable:     tx.Spendable,
	}
}

func convertUTXOInfo(inputs ...common.UTXOInfo) []commands.UTXOInfo {
	var result []commands.UTXOInfo
	for _, input := range inputs {
		result = append(result, commands.UTXOInfo{
			TxID: input.TxID,
			Vout: input.Vout,
		})
	}

	return result
}

func convertSpendInfo(inputs ...common.SpendInfo) ([]commands.SpendInfo, []commands.AssetInfo) {
	var result []commands.SpendInfo
	var assets []commands.AssetInfo
	for _, input := range inputs {
		result = append(result, commands.SpendInfo{
			Address: input.PublicAddress,
			Amount:  input.Amount,
		})

		if len(input.Asset.Hash) > 0 {
			// do not create output if no change
			if input.Asset.ChangeAmount > 0.0 {
				result = append(result, commands.SpendInfo{
					Address: input.Asset.ChangeAddress,
					Amount:  input.Asset.ChangeAmount,
				})
			}

			// Append output asset
			assets = append(assets, commands.AssetInfo{
				Address: input.PublicAddress,
				Asset:   input.Asset.Hash,
			})

			// do not create output if no change
			if input.Asset.ChangeAmount > 0.0 {
				assets = append(assets, commands.AssetInfo{
					Address: input.Asset.ChangeAddress,
					Asset:   input.Asset.Hash,
				})
			}
		}
	}

	if len(assets) == 0 {
		assets = nil
	}

	return result, assets
}

func getFundedPrivateKeys(ctx context.Context, client *rpc.Client, funded commands.FundedTransaction) ([]commands.Address, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.getFundedPrivateKeys")
	rawTx, err := commands.DecodeRawTransaction(ctx, client.Client, commands.Transaction(funded.Hex))
	if err != nil {
		log.WithError(err).
			Error("DecodeRawTransaction failed")
		return nil, ErrRPCError
	}

	decoded, err := commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		log.WithError(err).
			Error("ConvertToRawTransactionBitcoin failed")
		return nil, ErrRPCError
	}

	addressMap := make(map[commands.Address]commands.Address)
	for _, in := range decoded.Vin {
		txInfo, err := commands.GetTransaction(ctx, client.Client, in.Txid, true)
		if err != nil {
			log.WithError(err).
				Error("GetTransaction failed")
			return nil, ErrRPCError
		}

		addressMap[txInfo.Address] = txInfo.Address
		for _, d := range txInfo.Details {
			address := commands.Address(d.Address)
			addressMap[address] = address
		}
	}

	var addresses []commands.Address
	for _, address := range addressMap {
		addresses = append(addresses, address)
	}

	var privkeys []commands.Address
	for _, address := range addresses {
		privkey, err := commands.DumpPrivkey(ctx, client.Client, address)
		if err != nil {
			log.WithError(err).
				Error("DumpPrivkey failed")
			return nil, ErrRPCError
		}
		privkeys = append(privkeys, privkey)
	}

	return privkeys[:], nil
}
