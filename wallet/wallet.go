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

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

const (
	DefaultChainInterval time.Duration = 30 * time.Second

	ConfirmedBlockCount   = 3 // number of confirmation to consider transaction complete
	UnconfirmedBlockCount = 6 // number of confirmation to continue fetching addressInfos
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
			Info("Adding rpc client")
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

	go scheduledChainUpdate(ctx, chainsOptions.Names(), DefaultChainInterval)

	<-ctx.Done()
}

func (p *Wallet) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	ctx = handlers.ChainHandlerContext(ctx, p)

	const concurencyLevel = 4

	nats.SubscribeWorkers(ctx, common.CryptoAddressNextDepositSubject, concurencyLevel, handlers.OnCryptoAddressNextDeposit)

	log.Debug("Bank Wallet registered")
}

// common.Chain interface
func (p *Wallet) GetNewAddress(ctx context.Context, chain, account string) (string, error) {
	return GetNewAddress(ctx, chain, account)
}

func checkChainParams(interval time.Duration) time.Duration {
	if interval < time.Second {
		interval = DefaultChainInterval
	}
	return interval
}

func scheduledChainUpdate(ctx context.Context, chains []string, interval time.Duration) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.scheduledChainUpdate")
	db := appcontext.Database(ctx)

	interval = checkChainParams(interval)

	log = log.WithFields(logrus.Fields{
		"Interval": interval.String(),
	})

	log.Info("Start wallet Chain Scheduler")

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
			chain := model.String(state.Chain)
			allAddresses := make(AddressMap)

			// fetch unused addresses from database
			{
				unused, err := database.AllMempoolCryptoAddresses(db, chain)
				if err != nil {
					log.WithError(err).
						Error("Failed to AllMempoolCryptoAddresses")
					continue
				}

				appendAddresses(allAddresses, unused...)
			}

			// fetch unconfirmed addresses from database
			unconfirmed, err := database.AllUnconfirmedCryptoAddresses(db, chain, model.BlockID(state.Height-UnconfirmedBlockCount))
			{
				if err != nil {
					log.WithError(err).
						Error("Failed to AllUnconfirmedCryptoAddresses")
					continue
				}

				appendAddresses(allAddresses, unconfirmed...)
			}

			// fetch missing addresses from database
			{
				missing, err := database.FindCryptoAddressesNotInOperationInfo(db, chain)
				if err != nil {
					log.WithError(err).
						Error("Failed to FindCryptoAddressesNotInOperationInfo")
					continue
				}

				appendAddresses(allAddresses, missing...)
			}

			// fetch addresses with status received from database
			{
				received, err := database.FindCryptoAddressesByOperationInfoState(db, chain, model.String("received"))
				if err != nil {
					log.WithError(err).
						Error("Failed to FindCryptoAddressesByOperationInfoState")
					continue
				}

				appendAddresses(allAddresses, received...)
			}

			// create final addresses lists
			list, addresses := uniqueAddresses(allAddresses)

			log.WithField("List", list).
				Trace("Final publicAddresses")

			// Resquest chain
			infos, err := FetchChainAddressesInfo(ctx, state.Chain, state.Height, list...)
			if err != nil {
				log.WithError(err).
					Error("Failed to FetchChainAddressesInfo")
				continue
			}

			// local map for lookup cryptoAddresses from PublicAddress
			cryptoAddresses := make(map[model.String]*model.CryptoAddress)

			// update firstBlockId for NextDeposit
			for _, addr := range addresses {
				// search from
				for _, info := range infos {
					// the address is found
					if string(addr.PublicAddress) == info.PublicAddress {

						// update FirstBlockId
						firstBlockId := model.MemPoolBlockID // if returned FetchChainAddressesInfo, a tx exists at least in the mempool
						if info.Mined > 0 {
							firstBlockId = model.BlockID(info.Mined)
						}
						// skip if no changed
						if firstBlockId == addr.FirstBlockId {
							// store into local map
							cryptoAddresses[addr.PublicAddress] = &addr
							continue
						}

						// update FirstBlockId
						addr.FirstBlockId = firstBlockId

						// store into db
						cryptoAddress, err := database.AddOrUpdateCryptoAddress(db, addr)
						if err != nil {
							log.WithError(err).
								Error("Failed to AddOrUpdateCryptoAddress")
						}

						// update into local map
						cryptoAddresses[addr.PublicAddress] = &cryptoAddress
						break
					}
				}
			}

			// lookup for txid for account operations
			for _, addr := range addresses {
				// search from
				for _, info := range infos {
					// the address is found
					if string(addr.PublicAddress) == info.PublicAddress {

						// get assoiciated cryptoAddress from local map
						cryptoAddress := cryptoAddresses[addr.PublicAddress]
						if cryptoAddress == nil && cryptoAddress.ID == 0 {
							continue
						}

						// foreach transactions
						for _, transaction := range info.Transactions {
							// updateOperation
							err := updateOperation(ctx, cryptoAddress.ID, transaction)
							if err != nil {
								log.WithError(err).
									Error("Failed to updateOperation")
								continue
							}
						}
					}
				}
			}
		}
	}
}

func updateOperation(ctx context.Context, cryptoAddressID model.ID, transaction TransactionInfo) error {
	log := logger.Logger(ctx).WithField("Method", "Wallet.updateOperation")
	db := appcontext.Database(ctx)

	txID := model.TxID(transaction.TxID)

	log = log.WithFields(logrus.Fields{
		"CryptoAddressID": cryptoAddressID,
		"TxID":            txID,
	})

	// create OperationInfo and update OperationStatus
	err := db.Transaction(func(db bank.Database) error {
		operationInfo, err := database.GetOperationInfoByTxId(db, txID)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.WithError(err).
				Error("Failed to GetOperationInfoByTxId")
			return err
		}

		// operationInfo does not exists
		if operationInfo.ID == 0 {
			// create new OperationInfo
			info, err := database.AddOperationInfo(db, model.OperationInfo{
				CryptoAddressID: cryptoAddressID,
				TxID:            txID,
			})
			if err != nil {
				log.WithError(err).
					Error("Failed to AddOperationInfo")
				return err
			}

			// store result
			operationInfo = info
			log.WithField("OperationID", operationInfo.ID).
				Debug("OperationInfo created")
		}

		if operationInfo.ID == 0 {
			log.
				Error("Invalid operation ID")
			return database.ErrDatabaseError
		}

		log := log.WithField("operationInfoID", operationInfo.ID)

		// create or update OperationStatus
		operationState := "received"
		if transaction.Confirmations >= ConfirmedBlockCount {
			operationState = "confirmed"
		}

		// fetch OperationStatus if exists
		status, _ := database.GetOperationStatus(db, operationInfo.ID)
		if status.Accounted == "settled" {
			operationState = status.Accounted
		}

		// check if update is needed
		if status.State == operationState {
			return nil
		}

		// update state
		status, err = database.AddOrUpdateOperationStatus(db, model.OperationStatus{
			OperationInfoID: operationInfo.ID,
			State:           operationState,
			Accounted:       status.Accounted,
		})
		if err != nil {
			log.WithError(err).
				Error("Failed to AddOrUpdateOperationStatus")
			return err
		}

		log.WithField("OperationStatus", status.State).
			Debug("OperationStatus updated")

		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to perform database transaction")
		return err
	}

	return nil
}

type AddressMap map[string]model.CryptoAddress

func appendAddresses(allAddresses AddressMap, addresses ...model.CryptoAddress) {
	for _, addresse := range addresses {
		publicAddress := string(addresse.PublicAddress)
		if _, ok := allAddresses[publicAddress]; !ok {
			allAddresses[publicAddress] = addresse
		}
	}
}

func uniqueAddresses(allAddresses AddressMap) ([]string, []model.CryptoAddress) {
	// create final addresses lists
	var list []string                   // list for rpc call
	var addresses []model.CryptoAddress // list for operations update
	for publicAddress, addr := range allAddresses {
		list = append(list, publicAddress)
		addresses = append(addresses, addr)
	}

	return list, addresses
}
