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

type CryptoTransfert struct {
	Chain     string
	PublicKey string
}

type AccountTransferWithdraw struct {
	BatchMode string
	Source    AccountEntry
	Crypto    CryptoTransfert
}

type WithdrawInfo struct {
	WithdrawID uint64
	Timestamp  time.Time
	AccountID  uint64
	Amount     float64
	Chain      string
	PublicKey  string
	Status     string
}

type UserWithdraws struct {
	UserID    uint64
	Withdraws []WithdrawInfo
}

type BatchWithdraw struct {
	BatchID       uint64
	BankAccountID uint64
	Network       string
	Status        string
	TxID          string
	Withdraws     []WithdrawInfo
}

type BatchWithdraws struct {
	Network string
	Batches []BatchWithdraw
}

type BatchStatus struct {
	BatchID uint64
	Status  string
}

type BatchUpdate struct {
	BatchStatus
	TxID   string
	Height int
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

func (p *WithdrawInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *WithdrawInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *UserWithdraws) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *UserWithdraws) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchWithdraw) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchWithdraw) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchWithdraws) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchWithdraws) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchStatus) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchStatus) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func (p *BatchUpdate) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *BatchUpdate) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
