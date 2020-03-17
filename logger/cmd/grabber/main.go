// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Logger grabber fetch entries from redis
package main

import (
	"context"
	"flag"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"
)

type Args struct {
	App          appcontext.Options
	WithDatabase bool

	Redis    cache.RedisOptions
	Database database.Options
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "LogGrabber")
	flag.BoolVar(&args.WithDatabase, "withDatabase", false, "Store log to database (default false)")

	cache.OptionArgs(&args.Redis)
	database.OptionArgs(&args.Database)

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))

	if args.WithDatabase {
		ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))
		databaseLogger := logger.NewDatabaseLogger(ctx)
		ctx = appcontext.WithLogger(ctx, databaseLogger)
		defer databaseLogger.Close()
	}

	logger := logger.NewRedisLogger(ctx)
	// Start the log grabber
	logger.Grab(ctx)
}
