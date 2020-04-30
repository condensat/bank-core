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
	ErrInvalidOperationStatus = errors.New("Invalid OperationStatus")
)

// AddOrUpdateOperationStatus
func AddOrUpdateOperationStatus(db bank.Database, operation model.OperationStatus) (model.OperationStatus, error) {
	var result model.OperationStatus
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if operation.OperationInfoID == 0 {
		return result, ErrInvalidOperationInfoID
	}

	if len(operation.State) == 0 {
		return result, ErrInvalidOperationStatus
	}

	operation.LastUpdate = time.Now().UTC().Truncate(time.Second)

	err := gdb.
		Where(model.OperationStatus{
			OperationInfoID: operation.OperationInfoID,
		}).
		Assign(operation).
		FirstOrCreate(&result).Error

	return result, err
}

// GetOperationStatus
func GetOperationStatus(db bank.Database, operationInfoID model.ID) (model.OperationStatus, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.OperationStatus{}, errors.New("Invalid appcontext.Database")
	}

	if operationInfoID == 0 {
		return model.OperationStatus{}, ErrInvalidOperationInfoID
	}

	var result model.OperationStatus
	err := gdb.
		Where(model.OperationStatus{
			OperationInfoID: operationInfoID,
		}).
		First(&result).Error
	if err != nil {
		return model.OperationStatus{}, err
	}

	return result, nil
}
