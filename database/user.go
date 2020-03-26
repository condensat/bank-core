// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"

	"github.com/condensat/bank-core"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

func FindOrCreateUser(ctx context.Context, database bank.Database, name, email string) (*model.User, error) {
	switch db := database.DB().(type) {
	case *gorm.DB:

		result := model.User{
			Name:  name,
			Email: email,
		}
		err := db.
			Where("name = ?", name).
			Where("email = ?", email).
			FirstOrCreate(&result).Error

		return &result, err

	default:
		return nil, ErrInvalidDatabase
	}
}

func UserExists(ctx context.Context, userID uint64) bool {
	entry, err := FindUserById(ctx, userID)

	return err == nil && entry != nil && entry.ID > 0
}

func FindUserById(ctx context.Context, userID uint64) (*model.User, error) {
	db := appcontext.Database(ctx)

	switch db := db.DB().(type) {
	case *gorm.DB:

		var result model.User
		err := db.
			Where(&model.User{ID: userID}).
			First(&result).Error

		return &result, err

	default:
		return nil, ErrInvalidDatabase
	}
}
