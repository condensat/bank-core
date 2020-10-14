// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"context"
	"errors"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger/model"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type DatabaseLogger struct {
	database database.Context
}

func NewDatabaseLogger(ctx context.Context) *DatabaseLogger {
	database := appcontext.Database(ctx)
	db, ok := database.DB().(*gorm.DB)
	if !ok {
		log.
			Panic("Database is not gorm")
	}

	err := model.Migrate(db)
	if err != nil {
		log.
			WithError(err).
			Panic("Failed to migrate database")
	}

	ret := DatabaseLogger{
		database: database,
	}

	return &ret
}

func (p *DatabaseLogger) Close() {
	if db, ok := p.database.DB().(*gorm.DB); ok {
		db.Close()
	}
}

func (p *DatabaseLogger) CreateLogEntry(timestamp time.Time, app, level string, userID uint64, sessionID string, method, err, msg, data string) *model.LogEntry {
	return &model.LogEntry{
		Timestamp: timestamp.UTC().Round(time.Second),
		App:       app,
		Level:     level,
		UserID:    userID,
		SessionID: sessionID,
		Method:    method,
		Error:     err,
		Message:   msg,
		Data:      data,
	}
}

func (p *DatabaseLogger) AddLogEntries(entries []*model.LogEntry) error {
	if db, ok := p.database.DB().(*gorm.DB); ok {
		model.TxAddLogEntries(db, entries)
	}
	return errors.New("Invalid db")
}
