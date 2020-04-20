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
	ErrInvalidPublicAddress = errors.New("Invalid Public Address")
)

// AddOrUpdateCryptoAddress
func AddOrUpdateCryptoAddress(db bank.Database, address model.CryptoAddress) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if address.AccountID == 0 {
		return result, ErrInvalidAccountID
	}
	if len(address.PublicAddress) == 0 {
		return result, ErrInvalidPublicAddress
	}

	// set CreationDate for new entry
	if address.CreationDate == nil || address.CreationDate.IsZero() {
		creationDate := time.Now().UTC().Truncate(time.Second)
		address.CreationDate = &creationDate
	}
	// do not update CreationDate
	if address.AccountID != 0 {
		address.CreationDate = &time.Time{}
	}

	err := gdb.
		Where(model.CryptoAddress{
			AccountID: address.AccountID,
		}).
		Attrs(address).
		FirstOrCreate(&result).Error

	return result, err
}

func GetCryptoAddressByAccountID(db bank.Database, accountID model.AccountID) (model.CryptoAddress, error) {
	var result model.CryptoAddress

	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if accountID == 0 {
		return result, ErrInvalidAccountID
	}

	err := gdb.
		Where(model.CryptoAddress{
			AccountID: accountID,
		}).
		First(&result).Error

	return result, err
}
