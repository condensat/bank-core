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
	ErrInvalidSwapInfoID  = errors.New("Invalid SwapInfoID")
	ErrInvalidSwapPayload = errors.New("Invalid Swap Payload")
)

func AddSwapInfo(db bank.Database, swapID model.SwapID, status model.SwapStatus, payload model.Payload) (model.SwapInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SwapInfo{}, errors.New("Invalid appcontext.Database")
	}

	if swapID == 0 {
		return model.SwapInfo{}, ErrInvalidSwapID
	}
	if len(status) == 0 {
		return model.SwapInfo{}, ErrInvalidSwapType
	}
	if len(payload) == 0 {
		return model.SwapInfo{}, ErrInvalidSwapPayload
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.SwapInfo{
		Timestamp: timestamp,
		SwapID:    swapID,
		Status:    status,
		Payload:   payload,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.SwapInfo{}, err
	}

	return result, nil
}

type SwapInfos struct {
	Count  int
	Active int
}

func SwapssInfos(db bank.Database) (SwapInfos, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return SwapInfos{}, errors.New("Invalid appcontext.Database")
	}

	subQueryLast := gdb.Model(&model.SwapInfo{}).
		Select("MAX(id)").
		Group("swap_id").
		SubQuery()

	var totalSwaps int64
	err := gdb.Model(&model.SwapInfo{}).
		Where("id IN (?)", subQueryLast).
		Count(&totalSwaps).Error
	if err != nil {
		return SwapInfos{}, err
	}

	var activeSwaps int64
	err = gdb.Model(&model.SwapInfo{}).
		Where("swap_info.id IN (?)", subQueryLast).
		Where("status <> ?", "finalized").
		Count(&activeSwaps).Error
	if err != nil {
		return SwapInfos{}, err
	}

	return SwapInfos{
		Count:  int(totalSwaps),
		Active: int(activeSwaps),
	}, nil
}

func GetSwapInfo(db bank.Database, swapInfoID model.SwapInfoID) (model.SwapInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SwapInfo{}, errors.New("Invalid appcontext.Database")
	}

	if swapInfoID == 0 {
		return model.SwapInfo{}, ErrInvalidSwapInfoID
	}

	var result model.SwapInfo
	err := gdb.
		Where(&model.SwapInfo{ID: swapInfoID}).
		First(&result).Error
	if err != nil {
		return model.SwapInfo{}, err
	}

	return result, nil
}

func GetSwapInfoBySwapID(db bank.Database, swapID model.SwapID) (model.SwapInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SwapInfo{}, errors.New("Invalid appcontext.Database")
	}

	if swapID == 0 {
		return model.SwapInfo{}, ErrInvalidSwapID
	}

	var result model.SwapInfo
	err := gdb.
		Where(&model.SwapInfo{SwapID: swapID}).
		First(&result).Error
	if err != nil {
		return model.SwapInfo{}, err
	}

	return result, nil
}
