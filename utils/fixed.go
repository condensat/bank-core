// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package utils

import (
	"math"
	"math/big"
	"strconv"
)

const (
	DatabaseFloatingPrecision = 12
)

func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func RoundUnit(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}

func ToFixed(num float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}

	var f big.Float
	f.SetMode(big.AwayFromZero).SetFloat64(num)
	str := f.Text('f', precision)

	round, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return num // return original value
	}
	return round
}
