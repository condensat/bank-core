// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

type Model interface{}

// Database (GORM)
type DB interface{}

type Context interface {
	DB() DB

	Migrate(models []Model) error
	Transaction(txFunc func(tx Context) error) error
}
