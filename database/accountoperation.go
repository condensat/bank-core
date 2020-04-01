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
		return result, ErrInvalidDatabase
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
		return func() error {
			op, err := txApppendAccountOperation(db, operation)
			if err == nil {
				// write output result
				result = op
			}
			return err
		}()
	})

	// return result with error
	return result, err
}

func GetLastAccountOperation(db bank.Database, accountID model.AccountID) (model.AccountOperation, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return model.AccountOperation{}, ErrInvalidDatabase
	}

	if accountID == 0 {
		return model.AccountOperation{}, ErrInvalidAccountID
	}

	var result model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Last(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return model.AccountOperation{}, err
	}

	return result, err
}

func GeAccountHistory(db bank.Database, accountID model.AccountID) ([]model.AccountOperation, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return nil, ErrInvalidDatabase
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
		return nil, ErrInvalidDatabase
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

// ErrInvalidAccountOperation perform oerpation within a db transaction
func txApppendAccountOperation(db bank.Database, operation model.AccountOperation) (model.AccountOperation, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return model.AccountOperation{}, ErrInvalidDatabase
	}

	info, err := fetchAccountInfo(db, operation.AccountID)
	if err != nil {
		return model.AccountOperation{}, err
	}

	prepareNextOperation(&info, &operation)
	// pre-check operation with newupdated values
	if !operation.PreCheck() {
		return model.AccountOperation{}, ErrInvalidAccountOperation
	}

	// store operation
	err = gdb.Create(&operation).Error
	if err != nil {
		return model.AccountOperation{}, err
	}
	// check if operation is valid
	if !operation.IsValid() {
		return model.AccountOperation{}, ErrInvalidAccountOperation
	}

	return operation, nil
}

func fetchAccountInfo(db bank.Database, accountID model.AccountID) (AccountInfo, error) {
	// check for valid accountID
	if accountID == 0 {
		return AccountInfo{}, ErrInvalidAccountID
	}

	// get Account (for currency)
	account, err := GetAccountByID(db, accountID)
	if err != nil {
		return AccountInfo{}, ErrAccountNotFound
	}

	// check currency status
	curr, err := GetCurrencyByName(db, account.CurrencyName)
	if err != nil {
		return AccountInfo{}, ErrCurrencyNotFound
	}
	if !curr.IsAvailable() {
		return AccountInfo{}, ErrCurrencyNotAvailable
	}

	// check account status
	accountState, err := GetAccountStatusByAccountID(db, accountID)
	if err != nil {
		return AccountInfo{}, ErrAccountStateNotFound
	}
	if !accountState.State.Valid() {
		return AccountInfo{}, ErrInvalidAccountState
	}
	if accountState.State != model.AccountStatusNormal {
		return AccountInfo{}, ErrAccountIsDisabled
	}

	// update PrevID with last operation ID
	lastOperation, err := GetLastAccountOperation(db, accountID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return AccountInfo{}, err
	}

	return AccountInfo{
		Account:  account,
		Currency: curr,
		State:    accountState,
		Last:     lastOperation,
	}, nil
}

type AccountInfo struct {
	Account  model.Account
	Currency model.Currency
	State    model.AccountState
	Last     model.AccountOperation
}

func prepareNextOperation(info *AccountInfo, operation *model.AccountOperation) {
	// update PrevID with last operation ID
	operation.PrevID = info.Last.ID

	// compute Balance with last operation and new Amount
	*operation.Balance = *operation.Amount
	if info.Last.Balance != nil {
		*operation.Balance += *info.Last.Balance
	}

	// compute TotalLocked with last operation and new LockAmount
	*operation.TotalLocked = *operation.LockAmount
	if info.Last.TotalLocked != nil {
		*operation.TotalLocked += *info.Last.TotalLocked
	}

	// To fixed precision
	*operation.Amount = model.ToFixedFloat(*operation.Amount)
	*operation.Balance = model.ToFixedFloat(*operation.Balance)

	*operation.LockAmount = model.ToFixedFloat(*operation.LockAmount)
	*operation.TotalLocked = model.ToFixedFloat(*operation.TotalLocked)
}
