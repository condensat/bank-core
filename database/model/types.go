// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type ID uint64
type RefID ID

type String string
type Float float64

type Base58 String
type ZeroInt *int
type ZeroFloat *Float

type Model interface{}
