// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	syslog "log"
	"os"
	"time"

	"github.com/condensat/bank-core/logger/model"

	driver "github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type DatabaseLogger struct {
	db *gorm.DB
}

type DatabaseOptions struct {
	HostName      string
	Port          int
	User          string
	Password      string
	Database      string
	EnableLogging bool
}

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

	return db
}
func NewDatabaseLogger(options DatabaseOptions) *DatabaseLogger {
	db := connectMyql(options.HostName, options.Port, options.User, options.Password, options.Database)
	db.LogMode(options.EnableLogging)
	db.SetLogger(syslog.New(os.Stderr, "", 0))

	err := model.Migrate(db)
	if err != nil {
		log.
			WithError(err).
			Panic("Failed to migrate database")
	}

	ret := DatabaseLogger{
		db: db,
	}

	return &ret
}

func (p *DatabaseLogger) Close() {
	p.db.Close()
}

func (p *DatabaseLogger) CreateLogEntry(timestamp time.Time, app, level, msg, data string) *model.LogEntry {
	return &model.LogEntry{
		Timestamp: timestamp.UTC().Round(time.Second),
		App:       app,
		Level:     level,
		Msg:       msg,
		Data:      data,
	}
}

func (p *DatabaseLogger) AddLogEntries(entries []*model.LogEntry) error {
	return model.TxAddLogEntries(p.db, entries)
}
