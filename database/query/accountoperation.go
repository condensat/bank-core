// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"
	"time"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

const (
	HistoryMaxOperationCount = 1000
)

var (
	ErrInvalidAccountOperation = errors.New("Invalid Account Operation")
)

type AccountOperationPrevNext struct {
	model.AccountOperation
	Previous model.AccountOperationID
	Next     model.AccountOperationID
}

func AppendAccountOperation(db database.Context, operation model.AccountOperation) (model.AccountOperation, error) {
	result, err := AppendAccountOperationSlice(db, operation)
	if err != nil {
		return model.AccountOperation{}, err
	}
	if len(result) != 1 {
		return model.AccountOperation{}, ErrInvalidAccountOperation
	}
	return result[0], nil
}

func TxAppendAccountOperation(db database.Context, operation model.AccountOperation) (model.AccountOperation, error) {
	operation.Timestamp = operation.Timestamp.UTC().Truncate(time.Second)

	return txApppendAccountOperation(db, operation)
}

func AppendAccountOperationSlice(db database.Context, operations ...model.AccountOperation) ([]model.AccountOperation, error) {
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}

	var result []model.AccountOperation
	err := db.Transaction(func(db database.Context) error {
		var txErr error
		result, txErr = TxAppendAccountOperationSlice(db, operations...)
		return txErr
	})

	return result, err
}

func TxAppendAccountOperationSlice(db database.Context, operations ...model.AccountOperation) ([]model.AccountOperation, error) {
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}

	// pre-check all operations
	for _, operation := range operations {
		// check for valid accountID
		accountID := operation.AccountID
		if accountID == 0 {
			return nil, ErrInvalidAccountID
		}

		// UTC timestamp
		operation.Timestamp = operation.Timestamp.UTC().Truncate(time.Second)

		// pre-check operation with ids
		if !operation.PreCheck() {
			return nil, ErrInvalidAccountOperation
		}
	}

	// already within a db transaction
	var result []model.AccountOperation
	err := func(db database.Context) error {

		// append all operations in same transaction
		// returning error will cause rollback
		for _, operation := range operations {
			op, err := txApppendAccountOperation(db, operation)
			if err != nil {
				return err
			}
			result = append(result, op)
		}

		return nil
	}(db)

	// return result with error
	return result, err
}

func GetPreviousAccountOperation(db database.Context, accountID model.AccountID, operationID model.AccountOperationID) (model.AccountOperation, error) {
	if db == nil {
		return model.AccountOperation{}, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

	if accountID == 0 {
		return model.AccountOperation{}, ErrInvalidAccountID
	}
	if operationID == 0 {
		return model.AccountOperation{}, ErrInvalidAccountOperation
	}

	var result model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Where("id < ?", operationID).
		Order("id DESC", true).
		Take(&result).Error

	if err != nil {
		return model.AccountOperation{}, err
	}

	return result, err
}

func GetNextAccountOperation(db database.Context, accountID model.AccountID, operationID model.AccountOperationID) (model.AccountOperation, error) {
	if db == nil {
		return model.AccountOperation{}, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

	if accountID == 0 {
		return model.AccountOperation{}, ErrInvalidAccountID
	}
	if operationID == 0 {
		return model.AccountOperation{}, ErrInvalidAccountOperation
	}

	var result model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Where("id > ?", operationID).
		Order("id ASC", true).
		First(&result).Error

	if err != nil {
		return model.AccountOperation{}, err
	}

	return result, err
}

func GetLastAccountOperation(db database.Context, accountID model.AccountID) (model.AccountOperation, error) {
	if db == nil {
		return model.AccountOperation{}, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

	if accountID == 0 {
		return model.AccountOperation{}, ErrInvalidAccountID
	}

	var result model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Last(&result).Error

	if err != nil {
		return model.AccountOperation{}, err
	}

	return result, err
}

func GeAccountHistory(db database.Context, accountID model.AccountID) ([]model.AccountOperation, error) {
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

	if accountID == 0 {
		return nil, ErrInvalidAccountID
	}

	var list []*model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Order("id ASC").
		Limit(HistoryMaxOperationCount).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertAccountOperationList(list), nil
}

func GeAccountHistoryWithPrevNext(db database.Context, accountID model.AccountID) ([]AccountOperationPrevNext, error) {
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

	if accountID == 0 {
		return nil, ErrInvalidAccountID
	}

	var rows []*AccountOperationPrevNext
	const query = `
		SELECT *,
			(SELECT id FROM account_operation AS sub
					WHERE ops.account_id = sub.account_id AND sub.id < ops.id
					ORDER BY sub.id DESC
					LIMIT 1
			) AS previous,
			(SELECT id FROM account_operation AS sub
					WHERE ops.account_id = sub.account_id AND sub.id > ops.id
					ORDER BY sub.id ASC
					LIMIT 1
			) AS next
		FROM account_operation as ops WHERE account_id = ? ORDER BY id asc;`
	err := gdb.
		Raw(query, accountID).
		Find(&rows).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertAccountOperationPrevNextList(rows), nil
}

func GeAccountHistoryRange(db database.Context, accountID model.AccountID, from, to time.Time) ([]model.AccountOperation, error) {
	if db == nil {
		return nil, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

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
		Limit(HistoryMaxOperationCount).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertAccountOperationList(list), nil
}

func FindAccountOperationByReference(db database.Context, synchroneousType model.SynchroneousType, operationType model.OperationType, referenceID model.RefID) (model.AccountOperation, error) {
	if db == nil {
		return model.AccountOperation{}, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

	if len(synchroneousType) == 0 {
		return model.AccountOperation{}, model.ErrSynchroneousTypeInvalid
	}
	if len(operationType) == 0 {
		return model.AccountOperation{}, model.ErrOperationTypeInvalid
	}
	if referenceID == 0 {
		return model.AccountOperation{}, ErrInvalidReferenceID
	}

	var result model.AccountOperation
	err := gdb.
		Where(model.AccountOperation{
			SynchroneousType: synchroneousType,
			OperationType:    operationType,
			ReferenceID:      referenceID,
		}).
		Last(&result).Error

	if err != nil {
		return model.AccountOperation{}, err
	}

	return result, err
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

func convertAccountOperationPrevNextList(list []*AccountOperationPrevNext) []AccountOperationPrevNext {
	var result []AccountOperationPrevNext
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}

// ErrInvalidAccountOperation perform oerpation within a db transaction
func txApppendAccountOperation(db database.Context, operation model.AccountOperation) (model.AccountOperation, error) {
	if db == nil {
		return model.AccountOperation{}, database.ErrInvalidDatabase
	}
	gdb := db.DB().(*gorm.DB)

	if operation.OperationType != model.OperationTypeInit {

		info, err := fetchAccountInfo(db, operation.AccountID)
		if err != nil {
			return model.AccountOperation{}, err
		}
		prepareNextOperation(&info, &operation)
	}

	// pre-check operation with newupdated values
	if !operation.PreCheck() {
		return model.AccountOperation{}, ErrInvalidAccountOperation
	}

	// store operation
	err := gdb.Create(&operation).Error
	if err != nil {
		return model.AccountOperation{}, err
	}
	// check if operation is valid
	if !operation.IsValid() {
		return model.AccountOperation{}, ErrInvalidAccountOperation
	}

	return operation, nil
}

func fetchAccountInfo(db database.Context, accountID model.AccountID) (AccountInfo, error) {
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

	// fetch last operation
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
