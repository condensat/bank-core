// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"context"
)

const (
	MessagingKey = "Key.MessagingKey"
)

// WithMessaging returns a context with the messaging set
func WithMessaging(ctx context.Context, messaging Messaging) context.Context {
	return context.WithValue(ctx, MessagingKey, messaging)
}

func FromContext(ctx context.Context) Messaging {
	if ctxMessaging, ok := ctx.Value(MessagingKey).(Messaging); ok {
		return ctxMessaging
	}
	return nil
}
