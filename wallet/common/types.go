// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"github.com/condensat/bank-core/messaging"
)

type CryptoMode string

const (
	CryptoModeBitcoinCore CryptoMode = "bitcoin-core"
	CryptoModeCryptoSsm   CryptoMode = "crypto-ssm"
)

type ServerOptions struct {
	Protocol string
	HostName string
	Port     int
}

type CryptoAddress struct {
	CryptoAddressID  uint64
	Chain            string
	AccountID        uint64
	PublicAddress    string
	Unconfidential   string
	IgnoreAccounting bool
}

type SsmAddress struct {
	Chain       string
	Address     string
	PubKey      string
	BlindingKey string
}

type TransactionInfo struct {
	Chain         string
	Account       string
	Address       string
	Asset         string
	TxID          string
	Vout          int64
	Amount        float64
	Confirmations int64
	Spendable     bool
}

type AddressInfo struct {
	Chain          string
	PublicAddress  string
	Unconfidential string
	IsValid        bool
}

type UTXOInfo struct {
	TxID   string
	Vout   int
	Asset  string
	Amount float64
	Locked bool
}

type SpendAssetInfo struct {
	Hash          string
	ChangeAddress string
	ChangeAmount  float64
}

type SpendInfo struct {
	PublicAddress string
	Amount        float64
	// Asset optional
	Asset SpendAssetInfo
}

type SpendTx struct {
	TxID string
}

type WalletInfo struct {
	Chain  string
	Height int
	UTXOs  []UTXOInfo
}

type WalletStatus struct {
	Wallets []WalletInfo
}

func (p *CryptoAddress) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *CryptoAddress) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}

func (p *AddressInfo) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *AddressInfo) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}

func (p *WalletInfo) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *WalletInfo) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}

func (p *WalletStatus) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *WalletStatus) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}
