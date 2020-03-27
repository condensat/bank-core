// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"time"

	"github.com/condensat/bank-core"
)

type AccountInfo struct {
	AccountID uint64
	Currency  string
	Name      string
	Status    string
}

type UserAccounts struct {
	UserID uint64

	Accounts []AccountInfo
}

type AccountEntry struct {
	AccountID        uint64
	Currency         string
	OperationType    string
	SynchroneousType string

	Timestamp time.Time
	Label     string
	Amount    float64
	Balance   float64

	LockAmount  float64
	TotalLocked float64
}

type AccountHistory struct {
	AccountID uint64
	From      time.Time
	To        time.Time

	History []AccountEntry
}

func (p *UserAccounts) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *UserAccounts) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountHistory) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountHistory) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
