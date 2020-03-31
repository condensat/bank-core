// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidAccountOperation = errors.New("Invalid Account Operation")
)

func AppendAccountOperation(db bank.Database, operation model.AccountOperation) (model.AccountOperation, error) {
	var result model.AccountOperation
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// check for valid accountID
	accountID := operation.AccountID
	if accountID == 0 {
		return result, ErrInvalidAccountID
	}

	// UTC timestamp
	operation.Timestamp = operation.Timestamp.UTC().Truncate(time.Second)

	// pre-check operation with ids
	if !operation.PreCheck() {
		return result, ErrInvalidAccountOperation
	}

	// within a db transaction
	// returning error will cause rollback
	err := db.Transaction(func(db bank.Database) error {

		// get Account (for currency)
		account, err := GetAccountByID(db, accountID)
		if err != nil {
			return ErrAccountNotFound
		}

		// check currency status
		curr, err := GetCurrencyByName(db, account.CurrencyName)
		if err != nil {
			return ErrCurrencyNotFound
		}
		if !curr.IsAvailable() {
			return ErrCurrencyNotAvailable
		}

		// check account status
		accountState, err := GetAccountStatusByAccountID(db, accountID)
		if err != nil {
			return ErrAccountStateNotFound
		}
		if !accountState.State.Valid() {
			return ErrInvalidAccountState
		}
		if accountState.State != model.AccountStatusNormal {
			return ErrAccountIsDisabled
		}

		// update PrevID with last operation ID
		previousOperation, err := GetLastAccountOperation(db, accountID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		operation.PrevID = previousOperation.ID

		// compute Balance with last operation and new Amount
		*operation.Balance = *operation.Amount
		if previousOperation.Balance != nil {
			*operation.Balance += *previousOperation.Balance
		}

		// compute TotalLocked with last operation and new LockAmount
		*operation.TotalLocked = *operation.LockAmount
		if previousOperation.TotalLocked != nil {
			*operation.TotalLocked += *previousOperation.TotalLocked
		}

		// pre-check operation with newupdated values
		if !operation.PreCheck() {
			return ErrInvalidAccountOperation
		}

		// store operation
		gdb := getGormDB(db)
		if gdb != nil {
			err = gdb.Create(&operation).Error
			if err != nil {
				return err
			}
			// check if operation is valid
			if !operation.IsValid() {
				return ErrInvalidAccountOperation
			}

			// get result and commit transaction
			result = operation
		}

		return nil
	})

	// return result with error
	return result, err
}

func GetLastAccountOperation(db bank.Database, accountID model.AccountID) (model.AccountOperation, error) {
	var result model.AccountOperation

	gdb := getGormDB(db)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if accountID == 0 {
		return result, ErrInvalidAccountID
	}

	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Last(&result).Error

	return result, err
}

func GeAccountHistory(db bank.Database, accountID model.AccountID) ([]model.AccountOperation, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if accountID == 0 {
		return nil, ErrInvalidAccountID
	}

	var list []*model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertAccountOperationList(list), nil
}

func GeAccountHistoryRange(db bank.Database, accountID model.AccountID, from, to time.Time) ([]model.AccountOperation, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if accountID == 0 {
		return nil, ErrInvalidAccountID
	}

	from = from.UTC().Truncate(time.Second)
	to = to.UTC().Truncate(time.Second)

	if from.After(to) {
		from, to = to, from
	}

	var list []*model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Where("timestamp BETWEEN ? AND ?", from, to).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertAccountOperationList(list), nil
}

func convertAccountOperationList(list []*model.AccountOperation) []model.AccountOperation {
	var result []model.AccountOperation
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}
