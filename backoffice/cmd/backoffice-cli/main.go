// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/backoffice"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/logger/model"

	"github.com/jinzhu/gorm"
)

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BackOfficeCli")

	database.OptionArgs(&args.Database)

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))

	migrateDatabase(ctx)

	AccountsInfo(ctx)
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

func AccountsInfo(ctx context.Context) {
	db := appcontext.Database(ctx)

	gdb := db.DB().(*gorm.DB)
	logsInfo, err := model.LogsInfo(gdb)
	if err != nil {
		panic(err)
	}
	fmt.Printf("LogsInfo: %+v\n", logsInfo)

	userCount, err := database.UserCount(db)
	if err != nil {
		panic(err)
	}
	fmt.Println("UserCount", userCount)

	accountsInfo, err := database.AccountsInfos(db)
	if err != nil {
		panic(err)
	}
	for _, account := range accountsInfo.Accounts {
		fmt.Printf("Accounts: %+v\n", account)
	}
	fmt.Printf("\tCount: %d\n", accountsInfo.Count)
	fmt.Printf("\tActive: %d\n", accountsInfo.Active)

	batchsInfo, err := database.BatchsInfos(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Batchs: %+v\n", batchsInfo)

	withdrawsInfo, err := database.WithdrawsInfos(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Withdraws: %+v\n", withdrawsInfo)
}
