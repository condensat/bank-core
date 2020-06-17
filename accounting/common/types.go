// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"time"

	"github.com/condensat/bank-core"
)

type CurrencyType int

type CurrencyInfo struct {
	Name             string
	DisplayName      string
	Available        bool
	AutoCreate       bool
	Crypto           bool
	Type             CurrencyType
	Asset            bool
	DisplayPrecision uint
}

type CurrencyList struct {
	Currencies []CurrencyInfo
}

type AccountInfo struct {
	Timestamp   time.Time
	AccountID   uint64
	UserID      uint64
	Currency    CurrencyInfo
	Name        string
	Status      string
	Balance     float64
	TotalLocked float64
}

type AccountCreation struct {
	UserID uint64
	Info   AccountInfo
}

type UserAccounts struct {
	UserID uint64

	Accounts []AccountInfo
}

type AccountEntry struct {
	OperationID     uint64
	OperationPrevID uint64

	AccountID        uint64
	Currency         string
	ReferenceID      uint64
	OperationType    string
	SynchroneousType string

	Timestamp time.Time
	Label     string
	Amount    float64
	Balance   float64

	LockAmount  float64
	TotalLocked float64
}

type AccountTransfer struct {
	Source      AccountEntry
	Destination AccountEntry
}

type AccountHistory struct {
	AccountID   uint64
	DisplayName string
	Ticker      string
	From        time.Time
	To          time.Time

	Entries []AccountEntry
}

type AccountTransferWithdraw struct {
	BatchMode string
	Source    AccountEntry
}

func (p *CurrencyList) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CurrencyList) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *CurrencyInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *CurrencyInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountCreation) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountCreation) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *UserAccounts) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *UserAccounts) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountEntry) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountEntry) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountTransfer) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountTransfer) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountHistory) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountHistory) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *AccountTransferWithdraw) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *AccountTransferWithdraw) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
