// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/currency/rate"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/database"
)

type CurrencyRate struct {
	AppID string
}

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Database database.Options

	CurrencyRate CurrencyRate
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "CurrencyRateGrabber")

	cache.OptionArgs(&args.Redis)
	database.OptionArgs(&args.Database)

	flag.StringVar(&args.CurrencyRate.AppID, "appId", "", "OpenExchangeRates application Id")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))

	migrateDatabase(ctx)

	var rateGrabber rate.RateGrabber
	rateGrabber.Run(ctx, args.CurrencyRate.AppID)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(database.CurrencyRateModel())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate curencyRate models")
	}
}
