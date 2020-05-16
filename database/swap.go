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
	DefaultSwapValidity time.Duration = 24 * time.Hour
)

var (
	ErrInvalidSwapID     = errors.New("Invalid SwapID")
	ErrInvalidSwapType   = errors.New("Invalid Swap Type")
	ErrInvalidSwapAmount = errors.New("Invalid Amount")
)

func AddSwap(db bank.Database, swapType model.SwapType, cryptoAddressID model.CryptoAddressID, debitAsset model.AssetID, debitAmount model.Float, creditAsset model.AssetID, creditAmount model.Float) (model.Swap, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Swap{}, errors.New("Invalid appcontext.Database")
	}

	if len(swapType) == 0 {
		return model.Swap{}, ErrInvalidSwapType
	}
	if cryptoAddressID == 0 {
		return model.Swap{}, ErrInvalidCryptoAddressID
	}
	if debitAsset == 0 {
		return model.Swap{}, ErrInvalidAssetHash
	}
	if debitAmount <= 0.0 {
		return model.Swap{}, ErrInvalidSwapAmount
	}
	if creditAsset == 0 {
		return model.Swap{}, ErrInvalidAssetHash
	}
	if creditAmount <= 0.0 {
		return model.Swap{}, ErrInvalidSwapAmount
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.Swap{
		Timestamp:       timestamp,
		ValidUntil:      timestamp.Add(DefaultSwapValidity),
		Type:            swapType,
		CryptoAddressID: cryptoAddressID,
		DebitAsset:      debitAsset,
		DebitAmount:     debitAmount,
		CreditAsset:     creditAsset,
		CreditAmount:    creditAmount,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.Swap{}, err
	}

	return result, nil

}

func GetSwap(db bank.Database, swapID model.SwapID) (model.Swap, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Swap{}, errors.New("Invalid appcontext.Database")
	}

	if swapID == 0 {
		return model.Swap{}, ErrInvalidSwapID
	}

	var result model.Swap
	err := gdb.
		Where(&model.Swap{ID: swapID}).
		First(&result).Error
	if err != nil {
		return model.Swap{}, err
	}

	return result, nil
}

func GetSwapByCryptoAddressID(db bank.Database, cryptoAddressID model.CryptoAddressID) (model.Swap, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Swap{}, errors.New("Invalid appcontext.Database")
	}

	if cryptoAddressID == 0 {
		return model.Swap{}, ErrInvalidCryptoAddressID
	}

	var result model.Swap
	err := gdb.
		Where(&model.Swap{CryptoAddressID: cryptoAddressID}).
		First(&result).Error
	if err != nil {
		return model.Swap{}, err
	}

	return result, nil
}
