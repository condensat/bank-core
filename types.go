// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bank

import (
	"context"
	"time"

	logModel "github.com/condensat/bank-core/logger/model"

	"github.com/condensat/secureid"
)

type BankObject interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}

type ServerOptions struct {
	Protocol string
	HostName string
	Port     int
}

type Logger interface {
	Close()
	CreateLogEntry(timestamp time.Time, app, level string, userID uint64, sessionID string, method, err, msg, data string) *logModel.LogEntry
	AddLogEntries(entries []*logModel.LogEntry) error
}

// Messaging (Nats)
type NC interface{}

type MessageHandler func(ctx context.Context, subject string, message *Message) (*Message, error)
type Messaging interface {
	NC() NC

	SubscribeWorkers(ctx context.Context, subject string, workerCount int, handle MessageHandler)
	Subscribe(ctx context.Context, subject string, handle MessageHandler)

	Publish(ctx context.Context, subject string, message *Message) error

	Request(ctx context.Context, subject string, message *Message) (*Message, error)
	RequestWithTimeout(ctx context.Context, subject string, message *Message, timeout time.Duration) (*Message, error)
}

// Cache (Redis)
type RDB interface{}

type Cache interface {
	RDB() RDB
}

type Worker interface {
	Run(ctx context.Context, numWorkers int)
}

type SecureID interface {
	ToSecureID(context string, value secureid.Value) (secureid.SecureID, error)
	FromSecureID(context string, secureID secureid.SecureID) (secureid.Value, error)

	ToString(secureID secureid.SecureID) string
	Parse(secureID string) secureid.SecureID
}
