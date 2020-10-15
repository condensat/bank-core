// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"

	"github.com/condensat/bank-core/accounting"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/messaging"
	mprovider "github.com/condensat/bank-core/messaging/provider"
)

type Accounting struct {
	BankUser string
}

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     mprovider.NatsOptions
	Database database.Options

	Accounting Accounting
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankAccounting")

	cache.OptionArgs(&args.Redis)
	mprovider.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

	flag.StringVar(&args.Accounting.BankUser, "bankUser", "bank@condensat.tech", "Bank database email [bank@condensat.tech]")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = cache.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = messaging.WithMessaging(ctx, mprovider.NewNats(ctx, args.Nats))
	ctx = appcontext.WithDatabase(ctx, database.New(args.Database))

	migrateDatabase(ctx)
	createDefaultFeeInfo(ctx)

	bankUser := createBankAccounts(ctx, args.Accounting)

	var service accounting.Accounting
	service.Run(ctx, bankUser)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(accounting.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate accounting models")
	}
}

func createDefaultFeeInfo(ctx context.Context) {
	db := appcontext.Database(ctx)

	defaultFeeInfo := []model.FeeInfo{
		// Fiat
		{Currency: "CHF", Minimum: 0.5, Rate: model.DefaultFeeRate},
		{Currency: "EUR", Minimum: 0.5, Rate: model.DefaultFeeRate},

		// Crypto
		{Currency: "BTC", Minimum: 0.00001000, Rate: model.DefaultFeeRate},
		{Currency: "LBTC", Minimum: 0.00001000, Rate: model.DefaultFeeRate},
		{Currency: "TBTC", Minimum: 0.00001000, Rate: model.DefaultFeeRate},

		// Liquid Asset with quote
		{Currency: "USDt", Minimum: 0.5, Rate: model.DefaultFeeRate},
		{Currency: "LCAD", Minimum: 0.5, Rate: model.DefaultFeeRate},
	}

	for _, feeInfo := range defaultFeeInfo {
		// Check FeeInfo validity
		if !feeInfo.IsValid() {
			logger.Logger(ctx).
				WithField("Method", "main.createDefaultFeeInfo").
				WithField("FeeInfo", feeInfo).
				Panic("Invalid default feeInfo")
			continue
		}
		// Do not update feeInfo since it could have been updated since creation
		if query.FeeInfoExists(db, feeInfo.Currency) {
			continue
		}
		// create default FeeInfo
		_, err := query.AddOrUpdateFeeInfo(db, feeInfo)
		if err != nil {
			logger.Logger(ctx).WithError(err).
				WithField("Method", "main.createDefaultFeeInfo").
				WithField("FeeInfo", "feeInfo").
				Error("AddOrUpdateFeeInfo failed")
			continue
		}
	}
}

func createBankAccounts(ctx context.Context, accounting Accounting) model.User {
	db := appcontext.Database(ctx)

	ret := model.User{
		Name:  "Condensat Bank",
		Email: model.UserEmail(accounting.BankUser),
	}
	ret, err := query.FindOrCreateUser(db, ret)
	if err != nil {
		logger.Logger(ctx).
			WithError(err).
			WithField("UserID", ret.ID).
			WithField("Name", ret.Name).
			WithField("Email", ret.Email).
			Panic("Unable to FindOrCreateUser BankUser")
	}

	logger.Logger(ctx).
		WithError(err).
		WithField("UserID", ret.ID).
		WithField("Email", ret.Email).
		Info("BankUser")
	return ret
}
