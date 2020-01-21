// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"

	"github.com/condensat/bank-core"

	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

func FinddOrCreateUser(ctx context.Context, database bank.Database, name, email string) (*model.User, error) {
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
