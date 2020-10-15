// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	DefaultNonceExpire = time.Duration(3 * time.Minute)
)

var (
	ErrInvaliNonceName = errors.New("Invalid nonce name")
)

func ResetNonce(ctx context.Context, name string) error {
	if len(name) == 0 {
		return ErrInvaliNonceName
	}
	rdb := ToRedis(FromContext(ctx))

	_, err := rdb.Del(ctx, name).Result()

	return err
}

func Nonce(ctx context.Context, name string, nonce uint64) (uint64, error) {
	rdb := ToRedis(FromContext(ctx))
	if rdb == nil {
		return 0, ErrInternalError
	}
	if len(name) == 0 {
		return 0, ErrInvaliNonceName
	}
	prev, err := rdb.Get(ctx, name).Uint64()
	if err != nil && err.Error() != redis.Nil.Error() {
		return 0, err
	}

	if nonce > prev {
		_, err = rdb.Set(ctx, name, nonce, time.Hour).Result()
		if err != nil {
			return 0, err
		}
	}

	return prev, nil
}
