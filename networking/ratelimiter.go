// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package networking

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/networking/ratelimiter"

	"github.com/go-redis/redis_rate/v9"
)

var (
	ErrNetworkingInternalError = errors.New("Networking Internal Error")
)

var (
	ErrRateLimit = errors.New("RateLimitReached")

	DefaultPeerRequestPerSecond = ratelimiter.RateLimitInfo{
		Limit: redis_rate.Limit{
			Period: time.Second,
			Rate:   100,
			Burst:  100,
		},
		KeyPrefix: "PeerRequest",
	}
)

func RegisterRateLimiter(ctx context.Context, rateLimit ratelimiter.RateLimitInfo) context.Context {
	rateLimit.Burst = rateLimit.Rate // see rate_limite.PerSecond
	raterLimiter := ratelimiter.New(ctx, rateLimit)
	return context.WithValue(ctx, ratelimiter.MiddlewarePeerRequestPerSecondKey, raterLimiter)
}

// MiddlewarePeerRateLimiter return StatusTooManyRequests if rate limite is reached
func MiddlewarePeerRateLimiter(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := r.Context()

	switch limiter := ctx.Value(ratelimiter.MiddlewarePeerRequestPerSecondKey).(type) {
	case *ratelimiter.RateLimiter:

		if !limiter.Allowed(ctx, RequesterIP(r)) {
			log := logger.Logger(ctx).WithField("Method", "MiddlewarePeerRateLimiter")

			AppendRequestLog(log, r).
				WithError(ErrRateLimit).
				Warning("Too many requests")

			http.Error(rw, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next(rw, r)

	default:
		log := logger.Logger(ctx).WithField("Method", "MiddlewarePeerRateLimiter")

		AppendRequestLog(log, r).
			WithError(ErrNetworkingInternalError).
			Error("No limiter found")

		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
