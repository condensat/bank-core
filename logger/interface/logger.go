// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"time"

	"github.com/condensat/bank-core/logger/model"
)

type Logger interface {
	Close()
	CreateLogEntry(timestamp time.Time, app, level string, userID uint64, sessionID string, method, err, msg, data string) *model.LogEntry
	AddLogEntries(entries []*model.LogEntry) error
}
