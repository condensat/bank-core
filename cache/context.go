// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"errors"
)

const (
	CacheKey       = "Key.CacheKey"
	RedisLockerKey = "Key.RedisLockerKey"
)

var (
	ErrInternalError = errors.New("InternalError")
)

// WithCache returns a context with the messaging set
func WithCache(ctx context.Context, cache Cache) context.Context {
	return context.WithValue(ctx, CacheKey, cache)
}

func FromContext(ctx context.Context) Cache {
	if ctxCache, ok := ctx.Value(CacheKey).(Cache); ok {
		return ctxCache
	}
	return nil
}

func RedisMutexContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, RedisLockerKey, NewRedisMutex(ctx))
}

func RedisMutexFromContext(ctx context.Context) Mutex {
	switch redisMutex := ctx.Value(RedisLockerKey).(type) {
	case *RedisMutex:
		return redisMutex

	case Mutex:
		return redisMutex

	default:
		return nil
	}
}
