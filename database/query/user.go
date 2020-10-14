// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidUserID    = errors.New("Invalid UserID")
	ErrInvalidUserName  = errors.New("Invalid User Name")
	ErrInvalidUserEmail = errors.New("Invalid User Email")
)

func FindOrCreateUser(db database.Context, user model.User) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if len(user.Name) == 0 {
			return model.User{}, ErrInvalidUserName
		}

		if len(user.Email) == 0 {
			return model.User{}, ErrInvalidUserEmail
		}

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
		return model.User{}, database.ErrInvalidDatabase
	}
}

func UserExists(db database.Context, userID model.UserID) bool {
	entry, err := FindUserById(db, userID)

	return err == nil && entry.ID > 0
}

func UserCount(db database.Context) (int, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result int64
		err := gdb.
			Model(&model.User{}).
			Group("email").
			Count(&result).Error

		return int(result), err

	default:
		return 0, database.ErrInvalidDatabase
	}
}

func UserPagingCount(db database.Context, countByPage int) (int, error) {
	if countByPage <= 0 {
		countByPage = 1
	}

	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result int
		err := gdb.
			Model(&model.User{}).
			Count(&result).Error
		var partialPage int
		if result%countByPage > 0 {
			partialPage = 1
		}
		return result/countByPage + partialPage, err

	default:
		return 0, database.ErrInvalidDatabase
	}
}

func UserPage(db database.Context, userID model.UserID, countByPage int) ([]model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if userID < 1 {
			userID = 1
		}
		if countByPage <= 0 {
			countByPage = 1
		}

		var list []*model.User
		err := gdb.Model(&model.User{}).
			Select("id, email").
			Where("id >= ?", userID).
			Order("id ASC").
			Limit(countByPage).
			Find(&list).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		return convertUser(list), nil

	default:
		return nil, database.ErrInvalidDatabase
	}
}

func FindUserById(db database.Context, userID model.UserID) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if userID == 0 {
			return model.User{}, ErrInvalidUserID
		}

		var result model.User
		err := gdb.
			Where(&model.User{ID: userID}).
			First(&result).Error

		return result, err

	default:
		return model.User{}, database.ErrInvalidDatabase
	}
}

func FindUserByEmail(db database.Context, email model.UserEmail) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if len(email) == 0 {
			return model.User{}, ErrInvalidUserEmail
		}

		var result model.User
		err := gdb.
			Where(&model.User{Email: email}).
			First(&result).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return model.User{}, err
		}

		return result, nil

	default:
		return model.User{}, database.ErrInvalidDatabase
	}
}

func convertUser(list []*model.User) []model.User {
	var result []model.User
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}
