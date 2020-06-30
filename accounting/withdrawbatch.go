// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/sirupsen/logrus"
)

var (
	ErrProcessingWithdraw     = errors.New("Error Processing Withdraw")
	ErrProcessingWithdrawType = errors.New("Error Processing Withdraw Type")
)

func FetchCreatedWithdraws(ctx context.Context) ([]model.WithdrawTarget, error) {
	db := appcontext.Database(ctx)

	return database.GetLastWithdrawTargetByStatus(db, model.WithdrawStatusCreated)
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

		for _, data := range datas {
			// check if public key is valid
			if len(data.Data.PublicKey) == 0 {
				log.Error("Invalid Withdraw PublicKey")
				canceled = append(canceled, data.Withdraw.ID)
				continue
			}
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
			count, capacity, err := batchWithdrawCount(db, batchInfo.BatchID)
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

			// append batchIds to current batch
			err = database.AddWithdrawToBatch(db, batchInfo.BatchID, batchIDs...)
			if err != nil {
				canceled = append(canceled, batchIDs...)
				log.WithError(err).
					Error("Failed to AddWithdrawToBatch")
				return ErrProcessingWithdraw
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

func batchWithdrawCount(db bank.Database, batchID model.BatchID) (int, int, error) {
	batch, err := database.GetBatch(db, batchID)
	if err != nil {
		return 0, 0, err
	}
	withdraws, err := database.GetBatchWithdraws(db, batch.ID)
	if err != nil {
		return 0, 0, err
	}

	return len(withdraws), int(batch.Capacity), nil
}
