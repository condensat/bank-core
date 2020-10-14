// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"fmt"
	"io/ioutil"
	"strings"

	driver "github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func connectMyql(host string, port int, user, pass, dbname string) *gorm.DB {
	cfg := driver.Config{
		User:                 user,
		Passwd:               pass,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", host, port),
		DBName:               dbname,
		AllowNativePasswords: true,
		MultiStatements:      true,
		ParseTime:            true,
	}

	db, err := gorm.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.
			WithError(err).
			Panicln("Failed to open connection to database")
	}

	db.SingularTable(true)

	return db
}

func secretOrPassword(secret string) string {
	content, err := ioutil.ReadFile(secret)
	if err != nil {
		return secret
	}

	return strings.TrimRightFunc(string(content), func(c rune) bool {
		return c == '\r' || c == '\n'
	})
}
