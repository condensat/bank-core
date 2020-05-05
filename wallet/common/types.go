// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"github.com/condensat/bank-core"
)

type CryptoAddress struct {
	Chain          string
	AccountID      uint64
	PublicAddress  string
	Unconfidential string
}

type TransactionInfo struct {
	Chain         string
	Account       string
	Address       string
	TxID          string
	Amount        float64
	Confirmations int64
	Spendable     bool
}

type AddressInfo struct {
	Chain          string
	PublicAddress  string
	Unconfidential string
}

func (p *CryptoAddress) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CryptoAddress) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
