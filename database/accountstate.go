// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrAccountStateNotFound = errors.New("Account State Not Found")
	ErrInvalidAccountID     = errors.New("Invalid AccountID")
	ErrInvalidReferenceID   = errors.New("Invalid ReferenceID")
	ErrInvalidAccountState  = errors.New("Invalid Account State")
	ErrAccountIsDisabled    = errors.New("Account Is Disabled")
)

// AddOrUpdateAccountState
func AddOrUpdateAccountState(db bank.Database, accountSate model.AccountState) (model.AccountState, error) {
	var result model.AccountState
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if accountSate.AccountID == 0 {
		return result, ErrInvalidAccountID
	}

	if !accountSate.State.Valid() {
		return result, ErrInvalidAccountState
	}

	err := gdb.
		Where(model.AccountState{
			AccountID: accountSate.AccountID,
		}).
		Assign(accountSate).
		FirstOrCreate(&result).Error

	return result, err
}

func GetAccountStatusByAccountID(db bank.Database, accountID model.AccountID) (model.AccountState, error) {
	var result model.AccountState

	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if accountID == 0 {
		return result, ErrInvalidAccountID
	}

	err := gdb.
		Where(model.AccountState{
			AccountID: accountID,
		}).
		First(&result).Error

	return result, err
}
