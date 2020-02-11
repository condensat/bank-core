// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/api"
	"github.com/condensat/bank-core/api/ratelimiter"

	"github.com/condensat/bank-core/database"
)

type Api struct {
	Port              int
	CorsAllowedDomain string

	PeerRequestPerSecond ratelimiter.RateLimitInfo
	OpenSessionPerMinute ratelimiter.RateLimitInfo
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
	flag.StringVar(&args.Api.CorsAllowedDomain, "corsAllowedDomain", "condensat.space", "Cors Allowed Domain (default condensat.space)")

	args.Api.PeerRequestPerSecond = api.DefaultPeerRequestPerSecond
	flag.IntVar(&args.Api.PeerRequestPerSecond.Rate, "peerRateLimit", 100, "Rate limit rate, per second, per peer connection (default 100)")

	args.Api.OpenSessionPerMinute = api.DefaultOpenSessionPerMinute
	flag.IntVar(&args.Api.OpenSessionPerMinute.Rate, "sessionRateLimit", 10, "Open session limit rate, per minute, per user (default 10)")

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

	ctx = api.RegisterRateLimiter(ctx, args.Api.PeerRequestPerSecond)
	ctx = api.RegisterOpenSessionRateLimiter(ctx, args.Api.OpenSessionPerMinute)

	migrateDatabase(ctx)

	var api api.Api
	api.Run(ctx, args.Api.Port, corsAllowedOrigins(args.Api.CorsAllowedDomain))
}

func corsAllowedOrigins(corsAllowedDomain string) []string {
	// sub-domains wildcard
	return []string{fmt.Sprintf("https://%s.%s", "*", corsAllowedDomain)}
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(api.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate api models")
	}
}
