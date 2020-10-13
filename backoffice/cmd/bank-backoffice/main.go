// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/monitor"
	"github.com/condensat/bank-core/security/secureid"

	"github.com/condensat/bank-core/backoffice"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
)

type BackOffice struct {
	Port              int
	CorsAllowedDomain string

	SecureID string
}

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options

	BackOffice BackOffice
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BackOffice")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

	flag.IntVar(&args.BackOffice.Port, "port", 4242, "BankApi rpc port (default 4242)")
	flag.StringVar(&args.BackOffice.CorsAllowedDomain, "corsAllowedDomain", "condensat.space", "Cors Allowed Domain (default condensat.space)")

	flag.StringVar(&args.BackOffice.SecureID, "secureId", "secureid.json", "SecureID json file")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))
	ctx = appcontext.WithProcessusGrabber(ctx, monitor.NewProcessusGrabber(ctx, 15*time.Second))
	ctx = appcontext.WithSecureID(ctx, secureid.FromFile(args.BackOffice.SecureID))

	migrateDatabase(ctx)

	var backOffice backoffice.BackOffice
	backOffice.Run(ctx, args.BackOffice.Port, corsAllowedOrigins(args.BackOffice.CorsAllowedDomain))
}

func corsAllowedOrigins(corsAllowedDomain string) []string {
	// sub-domains wildcard
	return []string{fmt.Sprintf("https://%s.%s", "*", corsAllowedDomain)}
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(backoffice.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate backoffice models")
	}
}
