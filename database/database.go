// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	syslog "log"
	"os"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidDatabase = errors.New("Invalid database")
)

type Database struct {
	db *gorm.DB
}

func (p *Database) Transaction(txFunc func(tx bank.Database) error) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		return txFunc(&Database{db: tx})
	})
}

// NewDatabase create new mysql connection
// pannic if failed to connect
func NewDatabase(options Options) *Database {
	db := connectMyql(
		options.HostName, options.Port,
		options.User, appcontext.SecretOrPassword(options.Password),
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

func getGormDB(db bank.Database) *gorm.DB {
	if db == nil {
		return nil
	}

	switch gdb := db.DB().(type) {
	case *gorm.DB:
		return gdb

	default:
		return nil
	}
}

// zero allocation requests string for scope
const (
	reqEQ  = " = ?"
	reqGTE = " >= ?"
)
