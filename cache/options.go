// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"flag"
)

type RedisOptions struct {
	HostName string
	Port     int
}

func OptionArgs(args *RedisOptions) {
	if args == nil {
		panic("Invalid cache options")
	}

	flag.StringVar(&args.HostName, "redisHost", "cache", "Redis hostName (default 'cache')")
	flag.IntVar(&args.Port, "redisPort", 6379, "Redis port (default 6379)")
}
