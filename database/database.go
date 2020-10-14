// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"log"
	"os"

	"github.com/jinzhu/gorm"
)

const (
	DatabaseFloatingPrecision = 12
)

var (
	ErrInvalidDatabase = errors.New("Invalid database")
)

type Database struct {
	db *gorm.DB
}

// New create new mysql connection
// pannic if failed to connect
func New(options Options) Context {
	db := connectMyql(
		options.HostName, options.Port,
		options.User, secretOrPassword(options.Password),
		options.Database,
	)

	db.LogMode(options.EnableLogging)
	db.SetLogger(log.New(os.Stderr, "", 0))

	return &Database{
		db: db,
	}
}

// DB returns subsequent db connection
// see Context interface
func (d *Database) DB() DB {
	return d.db
}

func (p *Database) Migrate(models []Model) error {
	var interfaces []interface{}
	for _, model := range models {
		interfaces = append(interfaces, model)
	}
	return p.db.AutoMigrate(
		interfaces...,
	).Error
}

func (p *Database) Transaction(txFunc func(tx Context) error) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		return txFunc(&Database{db: tx})
	})
}
