// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"
	"strings"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/utils"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidAssetID   = errors.New("Invalid AssetID")
	ErrInvalidAssetHash = errors.New("Invalid AssetHash")
)

func AddAsset(db database.Context, assetHash model.AssetHash, currencyName model.CurrencyName) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, database.ErrInvalidDatabase
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

func AssetCount(db database.Context) (int64, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, database.ErrInvalidDatabase
	}

	var count int64
	err := gdb.Model(&model.Asset{}).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetAsset(db database.Context, assetID model.AssetID) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, database.ErrInvalidDatabase
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

func GetAssetByHash(db database.Context, assetHash model.AssetHash) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, database.ErrInvalidDatabase
	}

	if len(assetHash) == 0 {
		return model.Asset{}, ErrInvalidAssetHash
	}

	if utils.ContainEllipsis(string(assetHash)) {
		assetHash = getFullAssetHash(gdb, assetHash)
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

func GetAssetByCurrencyName(db database.Context, currencyName model.CurrencyName) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, database.ErrInvalidDatabase
	}

	if len(currencyName) == 0 {
		return model.Asset{}, ErrInvalidCurrencyName
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{CurrencyName: currencyName}).
		First(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil
}

func AssetHashExists(db database.Context, assetHash model.AssetHash) bool {
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

func getFullAssetHash(gdb *gorm.DB, assetHash model.AssetHash) model.AssetHash {
	tips := utils.SplitEllipsis(string(assetHash))
	if len(tips) != 2 {
		return assetHash
	}

	var result model.Asset
	err := gdb.
		Where("asset LIKE ?", strings.Join(tips, "%")).
		Where(&model.Asset{Hash: assetHash}).
		First(&result).Error
	if err != nil {
		return assetHash
	}

	return result.Hash
}
