// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/monitor/database/model"
)

func Models() []database.Model {
	return []database.Model{
		database.Model(new(model.ProcessInfo)),
	}
}
