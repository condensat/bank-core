// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/utils"
)

func ToFixedFloat(value Float) Float {
	fixed := utils.ToFixed(float64(value), database.DatabaseFloatingPrecision)
	return Float(fixed)
}
