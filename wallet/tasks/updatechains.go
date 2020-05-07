// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/accounting/client"
	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/cache"
	"github.com/condensat/bank-core/wallet/chain"
	wcommon "github.com/condensat/bank-core/wallet/common"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// UpdateChains
func UpdateChains(ctx context.Context, epoch time.Time, chains []string) {
	log := logger.Logger(ctx).WithField("Method", "task.ChainUpdate")

	chainsStates, err := chain.FetchChainsState(ctx, chains...)
	if err != nil {
		log.WithError(err).
			Error("Failed to FetchChainsState")
		return
	}

	log.WithFields(logrus.Fields{
		"Epoch": epoch.Truncate(time.Millisecond),
		"Count": len(chainsStates),
	}).Info("Chain state fetched")

	err = cache.UpdateRedisChain(ctx, chainsStates...)
	if err != nil {
		log.WithError(err).
			Error("Failed to UpdateRedisChain")
		return
	}

	for _, state := range chainsStates {
		updateChain(ctx, epoch, state)
	}
}

func updateChain(ctx context.Context, epoch time.Time, state chain.ChainState) {
	log := logger.Logger(ctx).WithField("Method", "task.ChainUpdate")
	db := appcontext.Database(ctx)

	list, addresses := fetchActiveAddresses(ctx, state)

	// Resquest chain
	infos, err := chain.FetchChainAddressesInfo(ctx, state, AddressInfoMinConfirmation, AddressInfoMaxConfirmation, list...)
	if err != nil {
		log.WithError(err).
			Error("Failed to FetchChainAddressesInfo")
		return
	}

	// local map for lookup cryptoAddresses from PublicAddress
	type CryptoTransaction struct {
		CryptoAddress model.CryptoAddress
		Transactions  []chain.TransactionInfo
		Currency      common.CurrencyInfo
	}
	cryptoTransactions := make(map[string]CryptoTransaction)

	// update firstBlockId for NextDeposit
	for _, info := range infos {
		for _, cryptoAddress := range addresses {
			// search for matching public address
			publicAddress := string(cryptoAddress.PublicAddress)
			if !matchPublicAddress(cryptoAddress, info.PublicAddress) {
				continue
			}

			// store into local map
			cryptoTransaction := CryptoTransaction{
				CryptoAddress: cryptoAddress,
				Transactions:  info.Transactions[:],
			}
			cryptoTransactions[publicAddress] = cryptoTransaction

			// update FirstBlockId
			firstBlockId := model.MemPoolBlockID // if returned FetchChainAddressesInfo, a tx exists at least in the mempool
			if info.Mined > 0 {
				firstBlockId = model.BlockID(info.Mined)
			}
			// skip if not changed
			if firstBlockId == cryptoAddress.FirstBlockId {
				continue
			}

			// update FirstBlockId
			cryptoTransaction.CryptoAddress.FirstBlockId = firstBlockId

			// store into db
			cryptoAddressUpdate, err := database.AddOrUpdateCryptoAddress(db, cryptoTransaction.CryptoAddress)
			if err != nil {
				log.WithError(err).
					Error("Failed to AddOrUpdateCryptoAddress")
			}

			// update cryptoAddress
			cryptoTransaction.CryptoAddress = cryptoAddressUpdate
			// update local map
			cryptoTransactions[publicAddress] = cryptoTransaction
			break
		}
	}

	// updateOperation transactions
	for _, cryptoTransaction := range cryptoTransactions {
		for _, transactions := range cryptoTransaction.Transactions {
			// ensure currency exists
			assetID, err := createAssetCurrency(ctx, transactions.Asset)
			if err != nil {
				log.WithError(err).
					Error("createAssetCurrency failed")
				continue
			}
			err = updateOperation(ctx, state, cryptoTransaction.CryptoAddress.ID, assetID, transactions)
			if err != nil {
				log.WithError(err).
					Error("Failed to updateOperation")
				continue
			}
		}
	}
}

func matchPublicAddress(crytoAddress model.CryptoAddress, address string) bool {
	if len(address) == 0 {
		return false
	}
	return string(crytoAddress.PublicAddress) == address || string(crytoAddress.Unconfidential) == address
}

func updateOperation(ctx context.Context, state chain.ChainState, cryptoAddressID model.CryptoAddressID, assetID model.AssetID, transaction chain.TransactionInfo) error {
	log := logger.Logger(ctx).WithField("Method", "Wallet.updateOperation")
	db := appcontext.Database(ctx)

	txID := model.TxID(transaction.TxID)
	vout := model.Vout(transaction.Vout)

	log = log.WithFields(logrus.Fields{
		"CryptoAddressID": cryptoAddressID,
		"Vout":            vout,
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
				Vout:            vout,
				AssetID:         assetID,
				Amount:          model.Float(transaction.Amount),
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
			// time to settle
			operationState = status.Accounted
		}

		// check if update is needed
		if status.State == operationState {
			return nil
		}

		client := wcommon.ChainClientFromContext(ctx, state.Chain)
		if client == nil {
			return chain.ErrChainClientNotFound
		}

		var unlockMode int
		switch operationState {
		case "received":
			// lock utxo
			unlockMode = 1

		case "settled":
			// unlock utxo
			unlockMode = 2
		}

		if unlockMode > 0 {
			// mode=1 lock (unlock=false)
			// mode=2 unlock (unlock=true)
			unlock := unlockMode == 2

			err = client.LockUnspent(ctx, unlock, wcommon.TransactionInfo{
				TxID: string(operationInfo.TxID),
				Vout: int64(operationInfo.Vout),
			})
			if err != nil {
				// non fatal
				// this can occure when state was not seen recieved
				log.WithError(err).
					WithField("Unlock", unlock).
					Warn("Failed to LockUnspent")
			}
			log.
				WithField("Unlock", unlock).
				Debug("LockUnspent done")
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

func addNewAddress(allAddresses AddressMap, addresses ...model.CryptoAddress) {
	for _, address := range addresses {
		publicAddress := string(address.PublicAddress)
		if _, ok := allAddresses[publicAddress]; !ok {
			allAddresses[publicAddress] = address
		}
	}
}

func createAssetCurrency(ctx context.Context, assetHash string) (model.AssetID, error) {
	log := logger.Logger(ctx).WithField("Method", "tasks.createAssetCurrency")
	db := appcontext.Database(ctx)

	// no asset, no error
	if len(assetHash) == 0 {
		return 0, nil
	}
	// 1 L-BTC = 1 L-BTC
	if assetHash == PolicyAssetLiquid {
		return 0, nil
	}

	// check if asset exists
	asset, err := database.GetAssetByHash(db, model.AssetHash(assetHash))
	if err == nil {
		return asset.ID, nil
	}

	// create asset & currency
	assetCount, err := database.AssetCount(db)
	if err != nil {
		log.WithError(err).
			Error("AssetCount failed")
		return 0, nil
	}

	// create CurrencyName
	currencyName := fmt.Sprintf("Li#%05d", assetCount+1)
	log = log.WithFields(logrus.Fields{
		"AssetHash":    assetHash,
		"CurrencyName": currencyName,
	})

	_, err = client.CurrencyCreate(ctx, currencyName, true, 0)
	if err != nil {
		log.WithError(err).
			Error("CurrencyCreate failed")
		return 0, err
	}

	// activate currency
	_, err = client.CurrencySetAvailable(ctx, currencyName, true)
	if err != nil {
		log.WithError(err).
			Error("CurrencySetAvailable failed")
		return 0, err
	}

	asset, err = database.AddAsset(db, model.AssetHash(assetHash), model.CurrencyName(currencyName))
	if err != nil {
		log.WithError(err).
			Error("AddAsset failed")
		return 0, err
	}

	log.
		WithField("AssetID", asset.ID).
		Debug("Asset Created")

	return asset.ID, nil
}

func fetchActiveAddresses(ctx context.Context, state chain.ChainState) ([]string, []model.CryptoAddress) {
	log := logger.Logger(ctx).WithField("Method", "task.fetchActiveAddresses")
	db := appcontext.Database(ctx)
	chainName := model.String(state.Chain)

	log = log.WithFields(logrus.Fields{
		"Chain":  state.Chain,
		"Height": state.Height,
	})

	// localMap for all unque addresses
	allAddresses := make(AddressMap)

	// fetch unused addresses from database
	{
		unused, err := database.AllUnusedCryptoAddresses(db, chainName)
		if err != nil {
			log.WithError(err).
				Error("Failed to AllUnusedCryptoAddresses")
			return nil, nil
		}

		addNewAddress(allAddresses, unused...)
	}

	// fetch mempool addresses from database
	{
		mempool, err := database.AllMempoolCryptoAddresses(db, chainName)
		if err != nil {
			log.WithError(err).
				Error("Failed to AllMempoolCryptoAddresses")
			return nil, nil
		}

		addNewAddress(allAddresses, mempool...)
	}

	// fetch unconfirmed addresses from database
	unconfirmed, err := database.AllUnconfirmedCryptoAddresses(db, chainName, model.BlockID(state.Height-UnconfirmedBlockCount))
	{
		if err != nil {
			log.WithError(err).
				Error("Failed to AllUnconfirmedCryptoAddresses")
			return nil, nil
		}

		addNewAddress(allAddresses, unconfirmed...)
	}

	// fetch missing addresses from database
	{
		missing, err := database.FindCryptoAddressesNotInOperationInfo(db, chainName)
		if err != nil {
			log.WithError(err).
				Error("Failed to FindCryptoAddressesNotInOperationInfo")
			return nil, nil
		}

		addNewAddress(allAddresses, missing...)
	}

	// fetch addresses with active state from database
	{
		activeStates := []model.String{
			"received",
			"confirmed",
		}
		received, err := database.FindCryptoAddressesByOperationInfoState(db, chainName, activeStates...)
		if err != nil {
			log.WithError(err).
				Error("Failed to FindCryptoAddressesByOperationInfoState")
			return nil, nil
		}

		addNewAddress(allAddresses, received...)
	}

	// create final addresses lists
	var result []string                 // addresses for rpc call
	var addresses []model.CryptoAddress // addresses for operations update
	for _, cryptoAddress := range allAddresses {
		address := string(cryptoAddress.PublicAddress)
		if len(cryptoAddress.Unconfidential) != 0 {
			// use unconfidential address for listunspent call
			address = string(cryptoAddress.Unconfidential)
		}
		result = append(result, address)
		addresses = append(addresses, cryptoAddress)
	}
	return result, addresses
}
