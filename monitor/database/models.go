// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	bank "github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/monitor/database/model"
)

func Models() []bank.Model {
	return []bank.Model{
		bank.Model(new(model.ProcessInfo)),
	}
}
