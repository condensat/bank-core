// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"github.com/condensat/bank-core"

	"github.com/go-redis/redis/v8"
)

func ToRedis(cache bank.Cache) *redis.Client {
	if cache == nil {
		return nil
	}
	rdb := cache.RDB()
	return rdb.(*redis.Client)
}
