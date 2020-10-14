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

var (
	ErrInvalidAssetInfo = errors.New("Invalid AssetInfo")
)

// AddOrUpdateAssetInfo
func AddOrUpdateAssetInfo(db database.Context, entry model.AssetInfo) (model.AssetInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.AssetInfo{}, database.ErrInvalidDatabase
	}

	if !entry.Valid() {
		return model.AssetInfo{}, ErrInvalidAssetInfo
	}

	entry.LastUpdate = time.Now().UTC().Truncate(time.Second)

	var result model.AssetInfo
	err := gdb.
		Where(model.AssetInfo{
			AssetID: entry.AssetID,
		}).
		Assign(entry).
		FirstOrCreate(&result).Error

	return result, err
}

// GetAssetInfo
func GetAssetInfo(db database.Context, assetID model.AssetID) (model.AssetInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.AssetInfo{}, database.ErrInvalidDatabase
	}

	if assetID == 0 {
		return model.AssetInfo{}, ErrInvalidAssetID
	}

	var result model.AssetInfo
	err := gdb.
		Where(&model.AssetInfo{AssetID: assetID}).
		First(&result).Error
	if err != nil {
		return model.AssetInfo{}, err
	}

	return result, nil
}
