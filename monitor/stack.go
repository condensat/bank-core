// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

import (
	"time"

	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/monitor/database/model"
)

type StackListService struct {
	Since       time.Duration
	Services    []string
	ProcessInfo []model.ProcessInfo
}

func (p *StackListService) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *StackListService) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}

type StackServiceHistory struct {
	AppName string
	From    time.Time
	To      time.Time
	Step    time.Duration
	Round   time.Duration
	History []model.ProcessInfo
}

func (p *StackServiceHistory) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *StackServiceHistory) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}
