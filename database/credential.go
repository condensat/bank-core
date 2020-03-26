// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"encoding/hex"
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

func HashEntry(entry model.Base58) model.Base58 {
	hash := utils.HashBytes([]byte(entry))
	return model.Base58(hex.EncodeToString(hash[:]))
}

func CreateOrUpdatedCredential(ctx context.Context, db bank.Database, credential model.Credential) (model.Credential, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		// perform a sha512 hex digest of login and password
		login := HashEntry(credential.LoginHash)
		password := HashEntry(credential.PasswordHash)
		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		passwordHash := security.SaltedHash(ctx, []byte(password))
		defer utils.Memzero(loginHash)
		defer utils.Memzero(passwordHash)

		var result model.Credential
		err := gdb.
			Where(&model.Credential{UserID: credential.UserID}).
			Assign(&model.Credential{
				LoginHash:    model.Base58(base58.Encode(loginHash, base58.BitcoinAlphabet)),
				PasswordHash: model.Base58(base58.Encode(passwordHash, base58.BitcoinAlphabet)),
				TOTPSecret:   credential.TOTPSecret,
			}).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.Credential{}, ErrInvalidDatabase
	}
}

func CheckCredential(ctx context.Context, db bank.Database, login, password model.Base58) (model.UserID, bool, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		// client should send a sha512 hex digest of the password
		// login = hashEntry(login)
		// password = hashEntry(password)

		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		defer utils.Memzero(loginHash)

		var cred model.Credential
		err := gdb.
			Where(&model.Credential{LoginHash: model.Base58(base58.Encode(loginHash, base58.BitcoinAlphabet))}).
			First(&cred).Error
		if err != nil {
			return 0, false, ErrDatabaseError
		}
		if cred.UserID == 0 {
			return 0, false, ErrUserNotFound
		}

		passwordHash, err := base58.Decode(string(cred.PasswordHash), base58.BitcoinAlphabet)
		defer utils.Memzero(passwordHash)
		if err != nil {
			return 0, false, ErrInvalidPasswordHash
		}

		return cred.UserID, security.SaltedHashVerify(ctx, []byte(password), passwordHash), nil

	default:
		return 0, false, ErrInvalidDatabase
	}
}
