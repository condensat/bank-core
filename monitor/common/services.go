// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"time"

	"github.com/condensat/bank-core"
)

type StackListService struct {
	Since       time.Duration
	Services    []string
	ProcessInfo []ProcessInfo
}

func (p *StackListService) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *StackListService) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

type StackServiceHistory struct {
	AppName string
	From    time.Time
	To      time.Time
	Step    time.Duration
	Round   time.Duration
	History []ProcessInfo
}

func (p *StackServiceHistory) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *StackServiceHistory) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
