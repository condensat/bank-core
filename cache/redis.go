// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"fmt"

	"github.com/condensat/bank-core"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	rdb *redis.Client
}

func NewRedis(ctx context.Context, options RedisOptions) *Redis {
	return &Redis{
		rdb: redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", options.HostName, options.Port),
		}),
	}
}

func (r *Redis) RDB() bank.RDB {
	return r.rdb
}
