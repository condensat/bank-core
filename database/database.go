// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	syslog "log"
	"os"

	"github.com/condensat/bank-core"

	"github.com/jinzhu/gorm"
)

type Database struct {
	db *gorm.DB
}

// NewDatabase create new mysql connection
// pannic if failed to connect
func NewDatabase(options Options) *Database {
	db := connectMyql(
		options.HostName, options.Port,
		options.User, options.Password,
		options.Database,
	)

	db.LogMode(options.EnableLogging)
	db.SetLogger(syslog.New(os.Stderr, "", 0))

	return &Database{
		db: db,
	}
}

// DB returns subsequent db connection
// see bank.Database interface
func (d *Database) DB() bank.DB {
	return d.db
}
