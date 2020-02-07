// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"

	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/security"
	"github.com/condensat/bank-core/security/utils"

	"github.com/jinzhu/gorm"
	"github.com/shengdoushi/base58"
)

var (
	ErrUserNotFound        = errors.New("User not found")
	ErrInvalidPasswordHash = errors.New("Invalid PasswordHash")
	ErrDatabaseError       = errors.New("Invalid PasswordHash")
)

func saltCredentials(ctx context.Context, login, password string) ([]byte, []byte) {
	loginHash := security.SaltedHash(ctx, []byte(login))
	passwordHash := security.SaltedHash(ctx, []byte(login+password))

	return loginHash, passwordHash
}

func CreateOrUpdatedCredential(ctx context.Context, database bank.Database, userID uint64, login, password, otpSecret string) (*model.Credential, error) {
	switch db := database.DB().(type) {
	case *gorm.DB:

		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		passwordHash := security.SaltedHash(ctx, []byte(password))
		defer utils.Memzero(loginHash)
		defer utils.Memzero(passwordHash)

		var cred model.Credential
		err := db.
			Where(&model.Credential{UserID: userID}).
			Assign(&model.Credential{
				LoginHash:    base58.Encode(loginHash, base58.BitcoinAlphabet),
				PasswordHash: base58.Encode(passwordHash, base58.BitcoinAlphabet),
				TOTPSecret:   otpSecret,
			}).
			FirstOrCreate(&cred).Error

		return &cred, err

	default:
		return nil, ErrInvalidDatabase
	}
}

func CheckCredential(ctx context.Context, database bank.Database, login, password string) (uint64, bool, error) {
	switch db := database.DB().(type) {
	case *gorm.DB:

		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		defer utils.Memzero(loginHash)

		var cred model.Credential
		err := db.
			Where(&model.Credential{LoginHash: base58.Encode(loginHash, base58.BitcoinAlphabet)}).
			First(&cred).Error
		if err != nil {
			return 0, false, ErrDatabaseError
		}
		if cred.UserID == 0 {
			return 0, false, ErrUserNotFound
		}

		passwordHash, err := base58.Decode(cred.PasswordHash, base58.BitcoinAlphabet)
		defer utils.Memzero(passwordHash)
		if err != nil {
			return 0, false, ErrInvalidPasswordHash
		}

		return cred.UserID, security.SaltedHashVerify(ctx, []byte(password), passwordHash), nil

	default:
		return 0, false, ErrInvalidDatabase
	}
}
