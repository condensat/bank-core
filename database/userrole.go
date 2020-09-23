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
	ErrInvalidRoleName = errors.New("Invalid RoleName")
)

func UserHasRole(db bank.Database, userID model.UserID, role model.RoleName) (bool, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return false, errors.New("Invalid appcontext.Database")
	}

	if userID == 0 {
		return false, ErrInvalidUserID
	}

	// all users has default role
	if role == model.RoleNameDefault {
		return true, nil
	}

	var userRole model.UserRole
	err := gdb.
		Where(model.UserRole{
			UserID: userID,
			Role:   role,
		}).First(&userRole).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	result := userRole.UserID == userID && userRole.Role == role

	return result, nil
}
