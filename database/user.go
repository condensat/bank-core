// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/condensat/bank-core"

	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

func FindOrCreateUser(db bank.Database, user model.User) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result model.User
		err := gdb.
			Where(model.User{
				Name:  user.Name,
				Email: user.Email,
			}).
			Assign(user).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.User{}, ErrInvalidDatabase
	}
}

func UserExists(db bank.Database, userID model.UserID) bool {
	entry, err := FindUserById(db, userID)

	return err == nil && entry.ID > 0
}

func FindUserById(db bank.Database, userID model.UserID) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result model.User
		err := gdb.
			Where(&model.User{ID: userID}).
			First(&result).Error

		return result, err

	default:
		return model.User{}, ErrInvalidDatabase
	}
}
