// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"github.com/condensat/bank-core"
)

type CryptoAddress struct {
	Chain         string
	AccountID     uint64
	PublicAddress string
}

func (p *CryptoAddress) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CryptoAddress) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
