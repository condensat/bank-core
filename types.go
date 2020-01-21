// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bank

import (
	"context"
	"time"

	"github.com/condensat/bank-core/logger/model"
)

type Key []byte

type PublicKey Key
type PrivateKey Key
type SharedKey Key

type Logger interface {
	Close()
	CreateLogEntry(timestamp time.Time, app, level, msg, data string) *model.LogEntry
	AddLogEntries(entries []*model.LogEntry) error
}

type MessageHandler func(ctx context.Context, subject string, message *Message) (*Message, error)

type Messaging interface {
	SubscribeWorkers(ctx context.Context, subject string, workerCount int, handle MessageHandler)
	Subscribe(ctx context.Context, subject string, handle MessageHandler)

	Request(ctx context.Context, subject string, message *Message) (*Message, error)
	RequestWithTimeout(ctx context.Context, subject string, message *Message, timeout time.Duration) (*Message, error)
}
