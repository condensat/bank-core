// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/monitor"

	"github.com/condensat/bank-core/cache"

	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/messaging/provider"
	mprovider "github.com/condensat/bank-core/messaging/provider"

	"github.com/condensat/bank-core/wallet"
)

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     mprovider.NatsOptions
	Database database.Options

	Wallet wallet.WalletOptions
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankWallet")

	cache.OptionArgs(&args.Redis)
	mprovider.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

	wallet.OptionArgs(&args.Wallet)

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithHasherWorker(ctx, args.App.Hasher)
	ctx = cache.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = messaging.WithMessaging(ctx, provider.NewNats(ctx, args.Nats))
	ctx = appcontext.WithDatabase(ctx, database.New(args.Database))
	ctx = appcontext.WithProcessusGrabber(ctx, monitor.NewProcessusGrabber(ctx, 15*time.Second))

	migrateDatabase(ctx)

	var service wallet.Wallet
	service.Run(ctx, args.Wallet)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(wallet.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate wallet models")
	}
}
