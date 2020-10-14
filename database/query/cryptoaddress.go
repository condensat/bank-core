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
	AllChains = "*"
)

var (
	ErrInvalidChain           = errors.New("Invalid Chain")
	ErrInvalidPublicAddress   = errors.New("Invalid Public Address")
	ErrInvalidCryptoAddressID = errors.New("Invalid CryptoAddress ID")
)

// AddOrUpdateCryptoAddress
func AddOrUpdateCryptoAddress(db database.Context, address model.CryptoAddress) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, database.ErrInvalidDatabase
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
		address.Unconfidential = ""
		address.Chain = ""
	}

	// search by id
	search := model.CryptoAddress{
		ID: address.ID,
	}
	// create entry
	if address.ID == 0 {
		search = model.CryptoAddress{
			AccountID:      address.AccountID,
			PublicAddress:  address.PublicAddress,
			Unconfidential: address.Unconfidential,
		}
	}

	err := gdb.
		Where(search).
		Assign(address).
		FirstOrCreate(&result).Error

	return result, err
}

func GetCryptoAddress(db database.Context, ID model.CryptoAddressID) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, database.ErrInvalidDatabase
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

func GetCryptoAddressWithPublicAddress(db database.Context, publicAddress model.String) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, database.ErrInvalidDatabase
	}

	if len(publicAddress) == 0 {
		return result, ErrInvalidPublicAddress
	}

	err := gdb.
		Where(model.CryptoAddress{
			PublicAddress: publicAddress,
		}).
		First(&result).Error

	return result, err
}

func GetCryptoAddressWithUnconfidential(db database.Context, unconfidential model.String) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, database.ErrInvalidDatabase
	}

	if len(unconfidential) == 0 {
		return result, ErrInvalidPublicAddress
	}

	err := gdb.
		Where(model.CryptoAddress{
			Unconfidential: unconfidential,
		}).
		First(&result).Error

	return result, err
}

func LastAccountCryptoAddress(db database.Context, accountID model.AccountID) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, database.ErrInvalidDatabase
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

func AllAccountCryptoAddresses(db database.Context, accountID model.AccountID) ([]model.CryptoAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return nil, database.ErrInvalidDatabase
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

func AllUnusedAccountCryptoAddresses(db database.Context, accountID model.AccountID) ([]model.CryptoAddress, error) {
	if accountID == 0 {
		return nil, ErrInvalidAccountID
	}
	return findUnusedCryptoAddresses(db, accountID, 0, AllChains)
}

func AllUnusedCryptoAddresses(db database.Context, chain model.String) ([]model.CryptoAddress, error) {
	return findUnusedCryptoAddresses(db, 0, 0, chain)
}

func AllMempoolCryptoAddresses(db database.Context, chain model.String) ([]model.CryptoAddress, error) {
	return findUnusedCryptoAddresses(db, 0, model.MemPoolBlockID, chain)
}

func AllUnconfirmedCryptoAddresses(db database.Context, chain model.String, afterBlock model.BlockID) ([]model.CryptoAddress, error) {
	return findUnusedCryptoAddresses(db, 0, afterBlock, chain)
}

func findUnusedCryptoAddresses(db database.Context, accountID model.AccountID, blockID model.BlockID, chain model.String) ([]model.CryptoAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return nil, database.ErrInvalidDatabase
	}

	if len(chain) == 0 {
		return nil, ErrInvalidChain
	}

	// support wildcard for all chains
	if chain == AllChains {
		chain = ""
	}

	var filter func(db *gorm.DB) *gorm.DB
	// filter for confirmed or unconfirmed
	if blockID > model.MemPoolBlockID {
		// filter confirmed
		filter = ScopeFirstBlockIDAfter(blockID)
	} else {
		// filter unconfirmed
		filter = ScopeFirstBlockIDExact(blockID)
	}

	var list []*model.CryptoAddress
	err := gdb.
		Where(model.CryptoAddress{
			AccountID: accountID,
			Chain:     chain,
		}).
		Scopes(filter).
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

// ScopeFirstBefore
func ScopeFirstBefore(blockID model.BlockID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqFirstBlockIDBefore(), blockID)
	}
}

// ScopeFirstBlockIDExact
func ScopeFirstBlockIDExact(blockID model.BlockID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqFirstBlockIDExact(), blockID)
	}
}

// ScopeFirstBlockIDAfter
func ScopeFirstBlockIDAfter(blockID model.BlockID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqFirstBlockIDAfter(), blockID)
	}
}

const (
	colAccountID        = "account_id"
	colPublicAddress    = "public_address"
	colUnconfidential   = "unconfidential"
	colChain            = "chain"
	colCreationDate     = "creation_date"
	colFirstBlockID     = "first_block_id"
	colIgnoreAccounting = "ignore_accounting"
)

func cryptoAddressColumnNames() []string {
	return []string{
		colID,
		colAccountID,
		colPublicAddress,
		colUnconfidential,
		colChain,
		colCreationDate,
		colFirstBlockID,
		colIgnoreAccounting,
	}
}

const ()

func reqFirstBlockIDBefore() string {
	var req [len(colFirstBlockID) + len(reqLTE)]byte
	off := 0
	off += copy(req[off:], colFirstBlockID)
	copy(req[off:], reqLTE)

	return string(req[:])
}

func reqFirstBlockIDExact() string {
	var req [len(colFirstBlockID) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colFirstBlockID)
	copy(req[off:], reqEQ)

	return string(req[:])
}

func reqFirstBlockIDAfter() string {
	var req [len(colFirstBlockID) + len(reqGTE)]byte
	off := 0
	off += copy(req[off:], colFirstBlockID)
	copy(req[off:], reqGTE)

	return string(req[:])
}
