// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"runtime"

	"github.com/condensat/bank-core/api"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/security"
	"github.com/shengdoushi/base58"

	"github.com/condensat/bank-core/database"
)

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankApi")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

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

	migrateDatabase(ctx)

	go testPasswordHash(ctx)

	api := new(api.Api)
	api.Run(ctx)
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

func testPasswordHash(ctx context.Context) {
	// create HasherWorker
	ctx = appcontext.WithHasherWorker(ctx, security.NewHasherWorker(ctx, 1, 256<<10, 4))

	// start HasherWorker
	numWorkers := runtime.NumCPU()
	go appcontext.HasherWorker(ctx).Run(ctx, numWorkers)

	var salt [16]byte
	_, _ = io.ReadFull(rand.Reader, salt[:])

	// simumlate clients
	for i := 0; i < 100; i++ {
		go func() {
			var password [32]byte
			_, _ = io.ReadFull(rand.Reader, password[:])

			key := security.SaltedHash(ctx, salt[:], password[:])
			fmt.Println(base58.Encode(key, base58.BitcoinAlphabet))

			if !security.SaltedHashVerify(ctx, salt[:], password[:], key) {
				logger.Logger(ctx).
					Panic("Failed to Verify SaltedHash")
			}
			logger.Logger(ctx).
				WithField("PasswordHash", base58.Encode(key, base58.BitcoinAlphabet)).
				Info("Password Hashed")

		}()
	}
}
