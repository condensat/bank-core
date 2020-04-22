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

const (
	AllChains = "*"
)

var (
	ErrInvalidChain           = errors.New("Invalid Chain")
	ErrInvalidPublicAddress   = errors.New("Invalid Public Address")
	ErrInvalidCryptoAddressID = errors.New("Invalid CryptoAddress ID")
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
	if len(address.Chain) == 0 || address.Chain == AllChains {
		return result, ErrInvalidChain
	}

	if address.ID == 0 {
		// set CreationDate for new entry
		timestamp := time.Now().UTC().Truncate(time.Second)
		address.CreationDate = &timestamp
	} else {
		// do not update non mutable fields
		address.CreationDate = nil
		address.PublicAddress = ""
		address.Chain = ""
	}

	// search by id
	search := model.CryptoAddress{
		ID: address.ID,
	}
	// create entry
	if address.ID == 0 {
		search = model.CryptoAddress{
			AccountID:     address.AccountID,
			PublicAddress: address.PublicAddress,
		}
	}

	err := gdb.
		Where(search).
		Assign(address).
		FirstOrCreate(&result).Error

	return result, err
}

func GetCryptoAddress(db bank.Database, ID model.ID) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return result, ErrInvalidCryptoAddressID
	}

	err := gdb.
		Where(model.CryptoAddress{
			ID: ID,
		}).
		First(&result).Error

	return result, err
}

func LastAccountCryptoAddress(db bank.Database, accountID model.AccountID) (model.CryptoAddress, error) {
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
		Last(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return result, err
	}

	return result, nil
}

func AllAccountCryptoAddresses(db bank.Database, accountID model.AccountID) ([]model.CryptoAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	if accountID == 0 {
		return nil, ErrInvalidAccountID
	}

	var list []*model.CryptoAddress
	err := gdb.
		Where(model.CryptoAddress{
			AccountID: accountID,
		}).
		Order("id ASC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return converCryptoAddressList(list), nil
}

func converCryptoAddressList(list []*model.CryptoAddress) []model.CryptoAddress {
	var result []model.CryptoAddress
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}
