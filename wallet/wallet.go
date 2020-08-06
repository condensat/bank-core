// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"context"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/bitcoin"
	"github.com/condensat/bank-core/wallet/chain"
	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/handlers"
	"github.com/condensat/bank-core/wallet/ssm"
	"github.com/condensat/bank-core/wallet/tasks"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

const (
	DefaultChainInterval      time.Duration = 30 * time.Second
	DefaultSsmInterval        time.Duration = 5 * time.Second
	DefaultOperationsInterval time.Duration = 5 * time.Second
	DefaultAssetInfoInterval  time.Duration = 30 * time.Second

	DefaultBatchInterval time.Duration = 1 * time.Minute

	ConfirmedBlockCount   = 3 // number of confirmation to consider transaction complete
	UnconfirmedBlockCount = 6 // number of confirmation to continue fetching addressInfos

	AddressInfoMinConfirmation = 0
	AddressInfoMaxConfirmation = 9999
)

type Wallet int

func (p *Wallet) Run(ctx context.Context, options WalletOptions) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.Run")

	// add RedisMutext to context
	ctx = cache.RedisMutexContext(ctx)
	ctx = common.CryptoModeContext(ctx, options.Mode)

	// load rpc clients configurations
	chainsOptions := loadChainsOptionsFromFile(options.FileName)

	// create all rpc clients
	for _, chainOption := range chainsOptions.Chains {
		log.WithField("Chain", chainOption.Chain).
			Info("Adding rpc client")
		ctx = common.ChainClientContext(ctx, chainOption.Chain, bitcoin.New(ctx, bitcoin.BitcoinOptions{
			ServerOptions: bank.ServerOptions{
				Protocol: "http",
				HostName: chainOption.HostName,
				Port:     chainOption.Port,
			},
			User: chainOption.User,
			Pass: chainOption.Pass,
		}))
	}

	ssmOptions := loadSsmOptionsFromFile(options.FileName)
	// create all ssm rpc clients
	for _, ssmDevice := range ssmOptions.Ssm.Devices {
		log.WithField("Device", ssmDevice.Name).
			Info("Adding ssm rpc client")

		ctx = common.SsmClientContext(ctx, ssmDevice.Name, ssm.NewWithTorEndpoint(ctx, ssmDevice.Endpoint))
	}

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Wallet Service started")

	go mainScheduler(ctx, chainsOptions.Names(), ssmOptions)

	<-ctx.Done()
}

func (p *Wallet) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	ctx = handlers.ChainHandlerContext(ctx, p)

	const concurencyLevel = 4

	nats.SubscribeWorkers(ctx, common.CryptoAddressNextDepositSubject, concurencyLevel, handlers.OnCryptoAddressNextDeposit)
	nats.SubscribeWorkers(ctx, common.CryptoAddressNewDepositSubject, concurencyLevel, handlers.OnCryptoAddressNewDeposit)
	nats.SubscribeWorkers(ctx, common.AddressInfoSubject, concurencyLevel, handlers.OnAddressInfo)

	log.Debug("Bank Wallet registered")
}

func mainScheduler(ctx context.Context, chains []string, ssmOptions SsmOptions) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.mainScheduler")

	taskChainUpdate := utils.Scheduler(ctx, DefaultChainInterval, 0)
	taskSsmPool := utils.Scheduler(ctx, DefaultSsmInterval, 0)
	taskOperationsUpdate := utils.Scheduler(ctx, DefaultOperationsInterval, 0)
	taskAssetInfoUpdate := utils.Scheduler(ctx, DefaultAssetInfoInterval, 0)
	taskBatchWithdraw := utils.Scheduler(ctx, DefaultBatchInterval, 0)

	// update once at startup
	tasks.UpdateAssetInfo(ctx, time.Now().UTC())

	var ssmInfo []tasks.SsmInfo
	for _, chain := range ssmOptions.Ssm.Chains {
		ssmInfo = append(ssmInfo, tasks.SsmInfo{
			Device:      chain.Device,
			Chain:       chain.Chain,
			Fingerprint: chain.Fingerprint,
		})
	}

	// Initialize SingleCalls nonce
	const singleCallPrefix = "bank.wallet."
	singleCalls := []string{
		singleCallPrefix + "UpdateChains",
		singleCallPrefix + "SsmPool",
		singleCallPrefix + "UpdateOperations",
		singleCallPrefix + "UpdateAssetInfo",
		singleCallPrefix + "BatchWithdraw",
	}
	for _, name := range singleCalls {
		err := cache.InitSingleCall(ctx, name)
		if err != nil {
			log.WithError(err).
				Panic("Failed to InitSingleCall")
		}
	}

	for {
		select {
		// update chains
		case epoch := <-taskChainUpdate:
			_ = cache.ExecuteSingleCall(ctx, singleCallPrefix+"UpdateChains",
				func(ctx context.Context) error {
					tasks.UpdateChains(ctx, epoch, chains)
					return nil
				})

		// ssm pool
		case epoch := <-taskSsmPool:
			_ = cache.ExecuteSingleCall(ctx, singleCallPrefix+"SsmPool",
				func(ctx context.Context) error {
					tasks.SsmPool(ctx, epoch, ssmInfo)
					return nil
				})

		// update operation
		case epoch := <-taskOperationsUpdate:
			_ = cache.ExecuteSingleCall(ctx, singleCallPrefix+"UpdateOperations",
				func(ctx context.Context) error {
					tasks.UpdateOperations(ctx, epoch, chains)
					return nil
				})

		// update assets
		case epoch := <-taskAssetInfoUpdate:
			_ = cache.ExecuteSingleCall(ctx, singleCallPrefix+"UpdateAssetInfo",
				func(ctx context.Context) error {
					tasks.UpdateAssetInfo(ctx, epoch)
					return nil
				})

		// batch withdraw
		case epoch := <-taskBatchWithdraw:
			_ = cache.ExecuteSingleCall(ctx, singleCallPrefix+"BatchWithdraw",
				func(ctx context.Context) error {
					tasks.BatchWithdraw(ctx, epoch, chains)
					return nil
				})

		case <-ctx.Done():
			log.Info("Daemon exited")
			return
		}
	}
}

// common.Chain interface
func (p *Wallet) GetNewAddress(ctx context.Context, chainName, account string) (string, error) {
	return chain.GetNewAddress(ctx, chainName, account)
}

func (p *Wallet) ImportAddress(ctx context.Context, chainName, account, address, pubkey string) error {
	return chain.ImportAddress(ctx, chainName, account, address, pubkey)
}

func (p *Wallet) GetAddressInfo(ctx context.Context, chainName, address string) (common.AddressInfo, error) {
	return chain.GetAddressInfo(ctx, chainName, address)
}
