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
	ErrInvalidAssetID   = errors.New("Invalid AssetID")
	ErrInvalidAssetHash = errors.New("Invalid AssetHash")
)

func AddAsset(db bank.Database, assetHash model.AssetHash, currencyName model.CurrencyName) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, errors.New("Invalid appcontext.Database")
	}

	if len(assetHash) == 0 {
		return model.Asset{}, ErrInvalidAssetHash
	}

	if len(currencyName) == 0 {
		return model.Asset{}, ErrInvalidCurrencyName
	}

	result := model.Asset{
		Hash:         assetHash,
		CurrencyName: currencyName,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil

}

func AssetCount(db bank.Database) (int64, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, errors.New("Invalid appcontext.Database")
	}

	var count int64
	err := gdb.Model(&model.Asset{}).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetAsset(db bank.Database, assetID model.AssetID) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, errors.New("Invalid appcontext.Database")
	}

	if assetID == 0 {
		return model.Asset{}, ErrInvalidAssetID
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{ID: assetID}).
		First(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil
}

func GetAssetByHash(db bank.Database, assetHash model.AssetHash) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, errors.New("Invalid appcontext.Database")
	}

	if len(assetHash) == 0 {
		return model.Asset{}, ErrInvalidAssetHash
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{Hash: assetHash}).
		First(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil
}

func AssetHashExists(db bank.Database, assetHash model.AssetHash) bool {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return false
	}

	if len(assetHash) == 0 {
		return false
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{Hash: assetHash}).
		First(&result).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return false
	}

	return true
}
