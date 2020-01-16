// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// simply push log entry to redis
package main

import (
	"context"
	"flag"
	"time"

	"github.com/condensat/bank-core/logger"
)

type Args struct {
	AppName  string
	LogLevel string
	Redis    logger.RedisOptions
}

func parseArgs() Args {
	var args Args
	flag.StringVar(&args.AppName, "appName", "LoggerExample", "Application Name")
	flag.StringVar(&args.LogLevel, "log", "warning", "Log level [trace, debug, info, warning, error]")

	flag.StringVar(&args.Redis.HostName, "redisHost", "localhost", "Redis hostName (default 'localhost')")
	flag.IntVar(&args.Redis.Port, "redisPort", 6379, "Redis port (default 6379)")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := logger.WithAppName(context.Background(), args.AppName)
	ctx = logger.WithWriter(ctx, logger.NewRedisLogger(args.Redis))
	ctx = logger.WithLogLevel(ctx, args.LogLevel)

	log := logger.Logger(ctx)
	for index := 0; index < 1024*10; index++ {
		log.
			WithField("Timestamp", time.Now().UTC().Round(time.Millisecond)).
			WithField("ID", index).
			Infof("Add log")

		if index%32 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
}
