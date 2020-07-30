// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// simply push log entry to redis
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/monitor/processus"

	"github.com/condensat/bank-core/accounting/client"
	"github.com/condensat/bank-core/accounting/common"

	"github.com/sirupsen/logrus"
)

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  messaging.NatsOptions
}

func parseArgs() Args {
	var args Args
	appcontext.OptionArgs(&args.App, "AccountManager")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)

	flag.Parse()

	return args
}

func exists(limit int, predicate func(i int) bool) bool {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return true
		}
	}
	return false
}

func createAndListAccount(ctx context.Context, currencies []common.CurrencyInfo, userID uint64) {
	log := logger.Logger(ctx).WithField("Method", "createAndListAccount")

	log = log.WithField("UserID", userID)

	// list user accounts
	userAccounts, err := client.AccountList(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("ListAccounts Failed")
		return
	}

	accounts := userAccounts.Accounts
	// create account for available currencies
	for _, currency := range currencies {
		if !currency.Available {
			continue
		}

		if exists(len(accounts), func(i int) bool {
			return accounts[i].Currency.Name == currency.Name
		}) {
			continue
		}
		account, err := client.AccountCreate(ctx, userID, currency.Name)
		if err != nil {
			log.WithError(err).
				Error("CreateAccount Failed")
			continue
		}
		log.WithField("Account", fmt.Sprintf("%+v", account)).
			Info("Account Created")

		// change account status to normal
		_, err = client.AccountSetStatus(ctx, account.Info.AccountID, "normal")
		if err != nil {
			log.WithError(err).
				Error("AccountSetStatus Failed")
			continue
		}

		_, err = client.AccountDepositSync(ctx, account.Info.AccountID, 42, 10.0, "First Deposit")
		if err != nil {
			log.WithError(err).
				Error("AccountDeposit Failed")
			continue
		}
	}

	for _, account := range accounts {

		// force write
		for i := 0; i < 1; i++ {
			client.AccountDepositSync(ctx, account.AccountID, 42, 0.1, "Batch Deposit")
			client.AccountDepositSync(ctx, account.AccountID, 42, -0.1, "Batch Deposit")
		}

		if account.AccountID > 4 {
			_, err = client.AccountTransfer(ctx, account.AccountID, 1+(account.AccountID-1)%4, 1337, account.Currency.Name, 0.01, "For weedcoder")
			if err != nil {
				log.WithError(err).
					Error("AccountTransfer Failed")
			}
		}

		to := time.Now()
		from := to.Add(-time.Hour)
		history, err := client.AccountHistory(ctx, account.AccountID, from, to)
		if err != nil {
			log.WithError(err).
				Error("AccountHistory Failed")
			return
		}

		log.WithFields(logrus.Fields{
			"AccountID": account.AccountID,
			"Count":     len(history.Entries),
		}).Infof("Account history")
	}
}

func CreateAccounts(ctx context.Context) {

	// list all currencies
	list, err := client.CurrencyList(ctx)
	if err != nil {
		panic(err)
	}

	var count int

	const userCount = 100
	for userID := 1; userID <= userCount; userID++ {
		createAndListAccount(ctx, list.Currencies, uint64(userID))
	}
	if userCount > 0 {
		return
	}

	start := time.Now()
	for i := 0; i < 10; i++ {
		// create users
		users := make([]uint64, 0, userCount)
		for userID := 1; userID <= userCount; userID++ {
			users = append(users, uint64(userID))
		}
		// randomize
		for i := len(users) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			users[i], users[j] = users[j], users[i]
		}
		batchSize := 128
		batches := make([][]uint64, 0, (len(users)+batchSize-1)/batchSize)
		for batchSize < len(users) {
			users, batches = users[batchSize:], append(batches, users[0:batchSize:batchSize])
		}
		batches = append(batches, users)

		for _, userIDs := range batches {

			var wait sync.WaitGroup
			for _, userID := range userIDs {
				wait.Add(1)

				go func(userID uint64) {
					defer wait.Done()

					createAndListAccount(ctx, list.Currencies, userID)
				}(uint64(userID))

				count++
			}
			wait.Wait()
		}
	}

	fmt.Printf("%d calls in %s\n", count, time.Since(start).Truncate(time.Millisecond))
}

func AccountTransferWithdraw(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "AccountTransferWithdraw")

	const accountID uint64 = 18
	log.WithField("AccountID", accountID)
	withdrawID, err := client.AccountTransferWithdrawCrypto(ctx,
		accountID, "TBTC", 0.00000300, "normal", "Test AccountTransferWithdraw",
		"bitcoin-testnet", "tb1qqjv0dec9vagycgwpchdkxsnapl9uy92dek4nau",
	)
	if err != nil {
		log.WithError(err).
			Error("AccountTransferWithdraw Failed")
		return
	}

	log.
		WithField("withdrawID", withdrawID).
		Info("AccountTransferWithdraw")
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))
	ctx = appcontext.WithProcessusGrabber(ctx, processus.NewGrabber(ctx, 15*time.Second))

	// CreateAccounts(ctx)
	AccountTransferWithdraw(ctx)
}
