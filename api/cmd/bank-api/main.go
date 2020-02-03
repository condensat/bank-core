// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"

	"github.com/condensat/bank-core/api"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/database"
)

type Api struct {
	Port int
}
type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options

	Api Api
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankApi")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

	flag.IntVar(&args.Api.Port, "port", 4242, "BankApi rpc port (default 4242)")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithHasherWorker(ctx, args.App.Hasher)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))

	migrateDatabase(ctx)

	var api api.Api
	api.Run(ctx, args.Api.Port)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(api.Models())
	if err != nil {
		logger.Logger(ctx).
			WithError(err).
			Panic("Failed to migrate api models")
	}
}
