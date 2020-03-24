// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package api

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidUserInfo        = errors.New("Invalid user info")
	ErrInvalidLoginOrPassword = errors.New("Invalid login or password")
	ErrInvalidEmail           = errors.New("Invalid email")
)

type UserInfo struct {
	Login,
	Password,
	Email string
	Roles []string
}

func ParseUserInfo(userInfo string) (*UserInfo, error) {
	toks := strings.Split(userInfo, ":")
	if len(toks) != 4 {
		return nil, ErrInvalidUserInfo
	}

	login := toks[0]
	password := toks[1]
	if len(login) == 0 || len(password) == 0 {
		return nil, ErrInvalidLoginOrPassword
	}

	email := toks[2]
	_, err := mail.ParseAddress(fmt.Sprintf("%s <%s>", login, email))
	if err != nil {
		return nil, ErrInvalidEmail
	}

	roles := strings.Split(toks[3], ",")
	if len(roles) == 0 {
		roles = append(roles, "user")
	}

	return &UserInfo{
		Login:    login,
		Password: password,
		Email:    email,
		Roles:    roles,
	}, nil
}

func scannerFromFileOrStdin(fileName string) (*bufio.Scanner, *os.File, error) {
	if len(fileName) == 0 || fileName == "-" {
		return bufio.NewScanner(os.Stdin), nil, nil
	} else {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, nil, err
		}
		return bufio.NewScanner(file), file, nil
	}
}

func FromUserInfoFile(ctx context.Context, fileName string) ([]*UserInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "api.FromUserInfoFile")
	scanner, file, err := scannerFromFileOrStdin(fileName)
	if err != nil {
		return nil, err
	}
	if file != nil {
		defer file.Close()
	}

	var result []*UserInfo
	for scanner.Scan() {
		userInfo, err := ParseUserInfo(scanner.Text())
		if err != nil {
			log.WithError(err).
				Error("Failed to ParseUserInfo")
			continue
		}
		result = append(result, userInfo)
	}
	return result[:], nil
}

func ImportUsers(ctx context.Context, userInfos ...*UserInfo) error {
	log := logger.Logger(ctx).WithField("Method", "api.ImportUsers")
	db := appcontext.Database(ctx)
	if db == nil {
		return errors.New("Invalid Database")
	}

	return db.Transaction(func(tx bank.Database) error {
		for _, userInfo := range userInfos {
			user, err := database.FindOrCreateUser(tx,
				userInfo.Login,
				userInfo.Email,
			)
			if err != nil {
				log.WithError(err).
					Error("Failed to FindOrCreateUser")
				continue
			}

			credential, err := database.CreateOrUpdatedCredential(ctx, tx,
				user.ID,
				userInfo.Login,
				userInfo.Password,
				"",
			)
			if err != nil {
				log.WithError(err).
					Error("Failed to CreateOrUpdatedCredential")
				continue
			}

			userID, verified, err := database.CheckCredential(ctx, tx,
				database.HashEntry(userInfo.Login),
				database.HashEntry(userInfo.Password),
			)
			if err != nil {
				log.WithError(err).
					Error("Failed to CheckCredential")
				continue
			}

			if !verified {
				log.Error("Not Verified")
				continue
			}

			if userID != user.ID {
				log.Error("Wrong UserID")
				continue
			}

			log.WithFields(logrus.Fields{
				"UserID":       userID,
				"LoginHash":    credential.LoginHash,
				"PasswordHash": credential.PasswordHash,
				"Verified":     verified,
			}).Info("User Imported")
		}
		return nil
	})
}
