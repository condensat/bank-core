// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"context"
	"time"
)

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
