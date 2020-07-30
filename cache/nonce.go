// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"errors"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/go-redis/redis/v7"
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
	rdb := ToRedis(appcontext.Cache(ctx))

	_, err := rdb.Del(name).Result()

	return err
}

func Nonce(ctx context.Context, name string, nonce uint64) (uint64, error) {
	rdb := ToRedis(appcontext.Cache(ctx))
	if rdb == nil {
		return 0, ErrInternalError
	}
	if len(name) == 0 {
		return 0, ErrInvaliNonceName
	}
	prev, err := rdb.Get(name).Uint64()
	if err != nil && err.Error() != redis.Nil.Error() {
		return 0, err
	}

	if nonce > prev {
		_, err = rdb.Set(name, nonce, time.Hour).Result()
		if err != nil {
			return 0, err
		}
	}

	return prev, nil
}
