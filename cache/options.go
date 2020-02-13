// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"flag"

	"github.com/condensat/bank-core"
)

type RedisOptions struct {
	bank.ServerOptions
}

func OptionArgs(args *RedisOptions) {
	if args == nil {
		panic("Invalid redis options")
	}

	flag.StringVar(&args.HostName, "redisHost", "cache", "Redis hostName (default 'cache')")
	flag.IntVar(&args.Port, "redisPort", 6379, "Redis port (default 6379)")
}
