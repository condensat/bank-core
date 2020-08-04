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
	ErrInvalidSsmAddressID     = errors.New("Invalid SsmAddressID")
	ErrInvalidSsmPublicAddress = errors.New("Invalid PublicAddress ID")
)

func AddSsmAddress(db bank.Database, address model.SsmAddress, info model.SsmAddressInfo) (model.SsmAddressID, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, errors.New("Invalid appcontext.Database")
	}

	if !address.IsValid() {
		return 0, errors.New("Invalid address")
	}
	info.SsmAddressID = model.SsmAddressID(1)
	if !info.IsValid() {
		return 0, errors.New("Invalid address info")
	}

	result := address
	err := gdb.Create(&result).Error
	if err != nil {
		return 0, err
	}

	info.SsmAddressID = result.ID
	if !info.IsValid() {
		return model.SsmAddressID(0), errors.New("Invalid address info")
	}
	err = gdb.Create(&info).Error
	if err != nil {
		return 0, err
	}

	_, err = UpdateSsmAddressState(db, result.ID, model.SsmAddressStatusUnused)
	if err != nil {
		return 0, nil
	}

	return result.ID, nil
}

func GetSsmAddress(db bank.Database, addressID model.SsmAddressID) (model.SsmAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddress{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddress{}, ErrInvalidSsmAddressID
	}

	var result model.SsmAddress
	err := gdb.
		Where(&model.SsmAddress{ID: addressID}).
		First(&result).Error
	if err != nil {
		return model.SsmAddress{}, err
	}

	return result, nil
}

func GetSsmAddressInfo(db bank.Database, addressID model.SsmAddressID) (model.SsmAddressInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddressInfo{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddressInfo{}, ErrInvalidSwapID
	}

	var result model.SsmAddressInfo
	err := gdb.
		Where(&model.SsmAddressInfo{SsmAddressID: addressID}).
		First(&result).Error
	if err != nil {
		return model.SsmAddressInfo{}, err
	}

	return result, nil
}

func GetSsmAddressByPublicAddress(db bank.Database, publicAddress model.SsmPublicAddress) (model.SsmAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddress{}, errors.New("Invalid appcontext.Database")
	}

	if len(publicAddress) == 0 {
		return model.SsmAddress{}, ErrInvalidCryptoAddressID
	}

	var result model.SsmAddress
	err := gdb.
		Where(&model.SsmAddress{PublicAddress: publicAddress}).
		First(&result).Error
	if err != nil {
		return model.SsmAddress{}, err
	}

	return result, nil
}

func GetSsmAddressState(db bank.Database, addressID model.SsmAddressID) (model.SsmAddressState, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddressState{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddressState{}, ErrInvalidSsmAddressID
	}

	var result model.SsmAddressState
	err := gdb.
		Where(&model.SsmAddressState{SsmAddressID: addressID}).
		Last(&result).Error
	if err != nil {
		return model.SsmAddressState{}, err
	}

	return result, nil
}

func UpdateSsmAddressState(db bank.Database, addressID model.SsmAddressID, status model.SsmAddressStatus) (model.SsmAddressState, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddressState{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddressState{}, ErrInvalidSsmAddressID
	}
	if len(status) == 0 {
		return model.SsmAddressState{}, ErrInvalidSsmAddressID
	}

	result := model.SsmAddressState{
		SsmAddressID: addressID,
		Timestamp:    time.Now().UTC().Truncate(time.Second),
		State:        status,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.SsmAddressState{}, err
	}

	return result, nil
}
