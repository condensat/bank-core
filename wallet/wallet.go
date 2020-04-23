// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"context"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/bitcoin"
	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/handlers"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

const (
	DefaultInterval time.Duration = 30 * time.Second
)

type Wallet int

func (p *Wallet) Run(ctx context.Context, options WalletOptions) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.Run")

	// add RedisMutext to context
	ctx = cache.RedisMutexContext(ctx)

	// load rpc clients configurations
	chainsOptions := loadChainsOptionsFromFile(options.FileName)

	// create all rpc clients
	for _, chainOption := range chainsOptions.Chains {
		log.WithField("Chain", chainOption.Chain).
			Warning("Adding rpc client")
		ctx = ChainClientContext(ctx, chainOption.Chain, bitcoin.New(ctx, bitcoin.BitcoinOptions{
			ServerOptions: bank.ServerOptions{
				HostName: chainOption.HostName,
				Port:     chainOption.Port,
			},
			User: chainOption.User,
			Pass: chainOption.Pass,
		}))
	}

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Wallet Service started")

	go p.scheduledUpdate(ctx, chainsOptions.Names(), DefaultInterval)

	<-ctx.Done()
}

func (p *Wallet) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 4

	nats.SubscribeWorkers(ctx, common.CryptoAddressNextDepositSubject, concurencyLevel, handlers.OnCryptoAddressNextDeposit)

	log.Debug("Bank Wallet registered")
}

func checkParams(interval time.Duration) time.Duration {
	if interval < time.Second {
		interval = DefaultInterval
	}
	return interval
}

func (p *Wallet) scheduledUpdate(ctx context.Context, chains []string, interval time.Duration) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.scheduledUpdate")
	db := appcontext.Database(ctx)

	interval = checkParams(interval)

	log = log.WithFields(logrus.Fields{
		"Interval": interval.String(),
	})

	log.Info("Start wallet Scheduler")

	for epoch := range utils.Scheduler(ctx, interval, 0) {
		chainsStates, err := FetchChainsState(ctx, chains...)
		if err != nil {
			log.WithError(err).
				Error("Failed to FetchChainsState")
			continue
		}

		log.WithFields(logrus.Fields{
			"Epoch": epoch.Truncate(time.Millisecond),
			"Count": len(chainsStates),
		}).Info("Chain state fetched")

		err = UpdateRedisChain(ctx, chainsStates)
		if err != nil {
			log.WithError(err).
				Error("Failed to UpdateRedisChain")
			continue
		}

		for _, state := range chainsStates {
			addresses, err := database.AllUnusedCryptoAddresses(db, model.String(state.Chain))
			if err != nil {
				log.WithError(err).
					Error("Failed to AllUnusedCryptoAddresses")
				continue
			}

			var list []string
			for _, addr := range addresses {
				list = append(list, string(addr.PublicAddress))
			}

			infos, err := FetchChainAddressesInfo(ctx, state.Chain, state.Height, list...)
			if err != nil {
				log.WithError(err).
					Error("Failed to FetchChainAddressesInfo")
				continue
			}

			for _, info := range infos {
				addr, err := database.GetCryptoAddressWithPublicAddress(db, model.String(info.PublicAddress))
				if err != nil {
					log.WithError(err).
						Error("Failed to GetCryptoAddressWithPublicAddress")
					continue
				}
				addr.FirstBlockId = model.BlockID(info.Mined)
				_, err = database.AddOrUpdateCryptoAddress(db, addr)
				if err != nil {
					log.WithError(err).
						Error("Failed to FetchChainAddressesInfo")
					continue
				}
			}
		}
	}
}
