// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/sirupsen/logrus"
)

var (
	ErrProcessingWithdraw     = errors.New("Error Processing Withdraw")
	ErrProcessingCanceling    = errors.New("Error Processing Canceling")
	ErrProcessingWithdrawType = errors.New("Error Processing Withdraw Type")
)

func FetchCreatedWithdraws(ctx context.Context) ([]model.WithdrawTarget, error) {
	db := appcontext.Database(ctx)

	return database.GetLastWithdrawTargetByStatus(db, model.WithdrawStatusCreated)
}

func FetchCancelingOperations(ctx context.Context) ([]model.AccountOperation, error) {
	db := appcontext.Database(ctx)

	return database.ListCancelingWithdrawsAccountOperations(db)
}

func ProcessWithdraws(ctx context.Context, withdraws []model.WithdrawTarget) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.ProcessWithdraws")

	byType := make(map[model.WithdrawTargetType][]model.WithdrawTarget)

	for _, withdraw := range withdraws {
		if _, ok := byType[withdraw.Type]; !ok {
			byType[withdraw.Type] = make([]model.WithdrawTarget, 0)
		}
		byType[withdraw.Type] = append(byType[withdraw.Type], withdraw)
	}

	for _, withdraws := range byType {
		err := processWithdraws(ctx, withdraws)
		if err != nil {
			log.WithError(err).Error("Fail to processWithdraws")
		}
	}

	return nil
}

func processWithdraws(ctx context.Context, withdraws []model.WithdrawTarget) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processWithdraws")
	db := appcontext.Database(ctx)

	if len(withdraws) == 0 {
		return nil
	}

	var datas []withdrawOnChainData
	wType := withdraws[0].Type

	switch wType {
	case model.WithdrawTargetOnChain:

		// fetch withdraw info from database
		for _, withdraw := range withdraws {
			// each withdraw should have same type
			if withdraw.Type != wType {
				log.WithFields(logrus.Fields{
					"RefType":      wType,
					"WithdrawType": withdraw.Type,
				}).Error("Wrong withdraw type")
				return ErrProcessingWithdrawType
			}

			// get withdraw
			w, err := database.GetWithdraw(db, withdraw.WithdrawID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetWithdraw")
				return err
			}
			// Get withdraw info history
			history, err := database.GetWithdrawHistory(db, withdraw.WithdrawID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetWithdrawHistory")
				return ErrProcessingWithdraw
			}
			// skip processed withdraw
			if len(history) != 1 || history[0].Status != model.WithdrawStatusCreated {
				log.Warn("Withdraw status is not created")
				continue
			}

			// get data
			data, err := withdraw.OnChainData()
			if err != nil {
				log.WithError(err).
					Error("Failed to get OnChainData")
				return ErrProcessingWithdraw
			}

			datas = append(datas, withdrawOnChainData{
				Withdraw: w,
				History:  history,
				Data:     data,
			})
		}

		return processWithdrawOnChain(ctx, datas)

	default:
		return ErrProcessingWithdrawType
	}
}

type withdrawOnChainData struct {
	Withdraw model.Withdraw
	History  []model.WithdrawInfo
	Data     model.WithdrawTargetOnChainData
}

func processWithdrawOnChain(ctx context.Context, datas []withdrawOnChainData) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processWithdrawOnChain")

	if len(datas) == 0 {
		log.Debug("Emtpy Withdraw data")
		return nil
	}

	// by chain withdraws map
	byChain := make(map[string][]withdrawOnChainData)

	for _, data := range datas {
		chain := data.Data.Chain
		if _, ok := byChain[chain]; !ok {
			byChain[chain] = make([]withdrawOnChainData, 0)
		}
		byChain[chain] = append(byChain[chain], data)
	}

	// process withdraw for same chain
	for chain, datas := range byChain {
		err := processWithdrawOnChainByNetwork(ctx, chain, datas)
		if err != nil {
			log.WithError(err).
				WithField("Chain", chain).
				Error("Failed to processWithdrawOnChainNetwork")
			continue
		}
	}

	return nil
}

func processWithdrawOnChainByNetwork(ctx context.Context, chain string, datas []withdrawOnChainData) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processWithdrawOnChainByNetwork")
	db := appcontext.Database(ctx)

	if len(chain) == 0 {
		log.Error("Invalid chain")
		return ErrProcessingWithdraw
	}
	if len(datas) == 0 {
		log.Debug("Emtpy Withdraw data")
		return nil
	}

	// Acquire Lock
	lock, err := cache.LockBatchNetwork(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock batchNetwork")
		return ErrProcessingWithdraw
	}
	defer lock.Unlock()

	var canceled []model.WithdrawID

	// within a db transaction
	err = db.Transaction(func(db bank.Database) error {

		var IDs []model.WithdrawID
		withdrawPubkeyMap := make(map[model.WithdrawID]string)
		for _, data := range datas {
			// check if public key is valid
			if len(data.Data.PublicKey) == 0 {
				log.Error("Invalid Withdraw PublicKey")
				canceled = append(canceled, data.Withdraw.ID)
				continue
			}

			// store withdrawID publicKey
			withdrawPubkeyMap[data.Withdraw.ID] = data.Data.PublicKey

			// check if withdraw amount is valid
			if data.Withdraw.Amount == nil || *data.Withdraw.Amount <= 0.0 {
				log.Error("Invalid Withdraw Amount")
				canceled = append(canceled, data.Withdraw.ID)
				continue
			}

			// change to status processing
			_, err := database.AddWithdrawInfo(db, data.Withdraw.ID, model.WithdrawStatusProcessing, "{}")
			if err != nil {
				log.WithError(err).
					Error("Failed to AddWithdrawInfo")

				canceled = append(canceled, IDs...)
				continue
			}

			IDs = append(IDs, data.Withdraw.ID)
		}

		var batchOffset int
		for len(IDs) > 0 {
			// create new batch regarding batchOffset
			batchInfo, err := findOrCreateBatchInfo(db, chain, batchOffset)
			if err != nil {
				log.WithError(err).
					Error("Failed to findOrCreateBatchInfo")
				return ErrProcessingWithdraw
			}

			// get capacity of current batch
			count, capacity, withdrawIDs, err := batchWithdrawCount(db, batchInfo.BatchID)
			if err != nil {
				log.WithError(err).
					Error("Failed to batchWithdrawCount")
				return ErrProcessingWithdraw
			}

			if count == capacity {
				// seek to next batch
				batchOffset++
				continue
			}

			addressMap := make(map[string]model.WithdrawID)
			for _, withdrawID := range withdrawIDs {
				wt, err := database.GetWithdrawTargetByWithdrawID(db, withdrawID)
				if err != nil {
					log.WithError(err).
						Error("GetWithdrawTargetByWithdrawID Failed")
					return ErrProcessingWithdraw
				}
				data, err := wt.OnChainData()
				if err != nil {
					log.WithError(err).
						Error("WithdrawTarget OnChainData Failed")
					return ErrProcessingWithdraw
				}
				// mark address as used
				addressMap[data.PublicKey] = withdrawID
			}

			// get all batch IDs
			batchIDs := IDs[:]
			{
				remaining := capacity - count
				if len(IDs) <= remaining {
					// all remaining fits in current batch
					IDs = nil // stop loop
				} else {
					// truncate IDs with remaining batch capacity
					batchIDs, IDs = IDs[:remaining], IDs[remaining:] // update batchIDs & IDs
				}
			}

			// find & remove witdraw from batch with same PublicKey
			batchCopy := make([]model.WithdrawID, len(batchIDs))
			copy(batchCopy, batchIDs)
			for i, batchID := range batchCopy {
				pubKey := withdrawPubkeyMap[batchID]
				if _, exists := addressMap[pubKey]; exists {
					batchIDs = removeWithdraw(batchIDs, i)            // remove from current batch
					IDs = append([]model.WithdrawID{batchID}, IDs...) // prepend for next batch
					continue
				}
			}

			// Add witdraws to batch
			if len(batchIDs) > 0 {
				// append batchIds to current batch
				err = database.AddWithdrawToBatch(db, batchInfo.BatchID, batchIDs...)
				if err != nil {
					canceled = append(canceled, batchIDs...)
					log.WithError(err).
						Error("Failed to AddWithdrawToBatch")
					return ErrProcessingWithdraw
				}
			}

			if len(IDs) > 0 {
				batchOffset++ // increment to get new batch in next step
			}
		}

		return nil
	})

	// update all canceled withdraws
	for _, ID := range canceled {
		_, err := database.AddWithdrawInfo(db, ID, model.WithdrawStatusCanceled, "{}")
		if err != nil {
			log.WithError(err).Error("failed to cancelWithdraw")
			continue
		}
	}

	if err != nil {
		return ErrProcessingWithdraw
	}

	return nil
}

func findOrCreateBatchInfo(db bank.Database, chain string, batchOffset int) (model.BatchInfo, error) {
	network := model.BatchNetwork(chain)
	batchCreated, err := database.GetLastBatchInfoByStatusAndNetwork(db, model.BatchStatusCreated, network)
	if err != nil {
		return model.BatchInfo{}, err
	}

	if len(batchCreated) > batchOffset {
		return batchCreated[batchOffset], nil
	}

	// create BatchInfo if not exists
	batch, err := database.AddBatch(db, network, model.BatchData(""))
	if err != nil {
		return model.BatchInfo{}, err
	}

	if err != nil {
		return model.BatchInfo{}, err
	}
	batchInfo, err := database.AddBatchInfo(db, batch.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}")
	if err != nil {
		return model.BatchInfo{}, err
	}

	return batchInfo, nil
}

func batchWithdrawCount(db bank.Database, batchID model.BatchID) (int, int, []model.WithdrawID, error) {
	batch, err := database.GetBatch(db, batchID)
	if err != nil {
		return 0, 0, nil, err
	}
	withdraws, err := database.GetBatchWithdraws(db, batch.ID)
	if err != nil {
		return 0, 0, nil, err
	}

	return len(withdraws), int(batch.Capacity), withdraws, nil
}

func accountRefund(ctx context.Context, db bank.Database, transfer common.AccountTransfer) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.accountRefund")

	log = log.WithFields(logrus.Fields{
		"SrcAccountID": transfer.Source.AccountID,
		"DstAccountID": transfer.Destination.AccountID,
		"Currency":     transfer.Source.Currency,
		"Amount":       transfer.Source.Amount,
	})

	// check operation type
	if model.OperationType(transfer.Destination.OperationType) != model.OperationTypeRefund {
		log.
			Error("OperationType is not refund")
		return common.AccountTransfer{}, database.ErrInvalidAccountOperation
	}
	// check for accounts
	if transfer.Source.AccountID == transfer.Destination.AccountID {
		log.
			Error("Can not transfer within same account")
		return common.AccountTransfer{}, database.ErrInvalidAccountOperation
	}

	// check for currencies match
	{
		// fetch source account from DB
		srcAccount, err := database.GetAccountByID(db, model.AccountID(transfer.Source.AccountID))
		if err != nil {
			log.WithError(err).
				Error("Failed to get srcAccount")
			return common.AccountTransfer{}, database.ErrInvalidAccountOperation
		}
		// fetch destination account from DB
		dstAccount, err := database.GetAccountByID(db, model.AccountID(transfer.Destination.AccountID))
		if err != nil {
			log.WithError(err).
				Error("Failed to get dstAccount")
			return common.AccountTransfer{}, database.ErrInvalidAccountOperation
		}
		// currency must match
		if srcAccount.CurrencyName != dstAccount.CurrencyName {
			log.WithFields(logrus.Fields{
				"SrcCurrency": srcAccount.CurrencyName,
				"DstCurrency": dstAccount.CurrencyName,
			}).Error("Can not transfer currencies")
			return common.AccountTransfer{}, database.ErrInvalidAccountOperation
		}
	}

	// Acquire Locks for source and destination accounts
	lockSource, err := cache.LockAccount(ctx, transfer.Source.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountTransfer{}, cache.ErrLockError
	}
	defer lockSource.Unlock()

	lockDestination, err := cache.LockAccount(ctx, transfer.Destination.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountTransfer{}, cache.ErrLockError
	}
	defer lockDestination.Unlock()

	// Prepare data
	transfer.Source.OperationType = transfer.Destination.OperationType
	transfer.Source.ReferenceID = transfer.Destination.ReferenceID
	transfer.Source.Timestamp = transfer.Destination.Timestamp
	transfer.Source.Label = transfer.Destination.Label

	transfer.Source.SynchroneousType = transfer.Destination.SynchroneousType
	transfer.Source.Amount = -transfer.Destination.Amount // do not create money
	transfer.Source.LockAmount = transfer.Source.Amount   // unlock funds
	transfer.Destination.LockAmount = 0.0

	// Store operations
	var operations []model.AccountOperation
	opSrc, err := database.TxAppendAccountOperation(db, common.ConvertEntryToOperation(transfer.Source))
	if err != nil {
		log.WithError(err).
			Error("Failed to AppendAccountOperationSlice")
		return common.AccountTransfer{}, err
	}
	operations = append(operations, opSrc)
	opDst, err := database.TxAppendAccountOperation(db, common.ConvertEntryToOperation(transfer.Destination))
	if err != nil {
		log.WithError(err).
			Error("Failed to AppendAccountOperationSlice")
		return common.AccountTransfer{}, err
	}
	operations = append(operations, opDst)

	// response should contains 2 operations
	if len(operations) != 2 {
		log.
			Error("Invalid operations count")
		return common.AccountTransfer{}, database.ErrInvalidAccountOperation
	}

	source := operations[0]
	destination := operations[1]
	log.Trace("Account transfer")

	return common.AccountTransfer{
		Source:      common.ConvertOperationToEntry(source, "N/A"),
		Destination: common.ConvertOperationToEntry(destination, "N/A"),
	}, nil
}

func removeWithdraw(s []model.WithdrawID, i int) []model.WithdrawID {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
