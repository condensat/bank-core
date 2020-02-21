// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

import (
	"github.com/condensat/bank-core"
)

func (p *ProcessInfo) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *ProcessInfo) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}
