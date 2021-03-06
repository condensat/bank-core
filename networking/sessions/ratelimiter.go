// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"context"
	"fmt"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/networking/ratelimiter"
)

func OpenSessionAllowed(ctx context.Context, userID uint64) bool {
	switch limiter := ctx.Value(ratelimiter.OpenSessionPerMinuteKey).(type) {
	case *ratelimiter.RateLimiter:

		return limiter.Allowed(ctx, fmt.Sprintf("UserID:%d", userID))

	default:
		logger.Logger(ctx).WithField("Method", "OpenSessionAllowed").
			Error("Failed to get OpenSession Limiter")
		return false
	}
}
