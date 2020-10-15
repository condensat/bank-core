// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/backoffice"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"

	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	mprovider "github.com/condensat/bank-core/messaging/provider"

	"github.com/condensat/bank-core/logger"
	logmodel "github.com/condensat/bank-core/logger/model"

	"github.com/jinzhu/gorm"
)

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     mprovider.NatsOptions
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
	ctx = appcontext.WithDatabase(ctx, database.New(args.Database))

	migrateDatabase(ctx)

	AccountsInfo(ctx)
	UsersInfo(ctx)
	DepositList(ctx)
	BatchList(ctx)
	WithdrawList(ctx)
	SwapList(ctx)
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
	logsInfo, err := logmodel.LogsInfo(gdb)
	if err != nil {
		panic(err)
	}
	fmt.Printf("LogsInfo: %+v\n", logsInfo)

	userCount, err := query.UserCount(db)
	if err != nil {
		panic(err)
	}
	fmt.Println("UserCount", userCount)

	accountsInfo, err := query.AccountsInfos(db)
	if err != nil {
		panic(err)
	}
	for _, account := range accountsInfo.Accounts {
		fmt.Printf("Accounts: %+v\n", account)
	}
	fmt.Printf("\tCount: %d\n", accountsInfo.Count)
	fmt.Printf("\tActive: %d\n", accountsInfo.Active)

	batchsInfo, err := query.BatchsInfos(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Batchs: %+v\n", batchsInfo)

	withdrawsInfo, err := query.WithdrawsInfos(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Withdraws: %+v\n", withdrawsInfo)
}

func UsersInfo(ctx context.Context) {
	db := appcontext.Database(ctx)

	pages, err := query.UserPagingCount(db, 5)
	if err != nil {
		panic(err)
	}
	fmt.Printf("User Pages: %d\n", pages)

	var start string
	for page := 0; page < pages; page++ {
		startID, _ := strconv.Atoi(start)
		userPage, err := query.UserPage(db, model.UserID(startID), 5)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Page %d: Users %+v\n", page, userPage)
		if len(userPage) == 0 {
			break
		}
		start = fmt.Sprintf("%d", int(userPage[len(userPage)-1].ID)+1)
	}
}

func DepositList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := query.DepositPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Deposit Page: %d\n", page)
}

func BatchList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := query.BatchPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Batch Page: %d\n", page)
}

func WithdrawList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := query.WithdrawPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Withdraw Page: %d\n", page)
}

func SwapList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := query.SwapPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Swap Page: %d\n", page)
}
