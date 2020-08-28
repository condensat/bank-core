// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"context"

	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/database/model"
)

const (
	BankUserKey = "Accounting.BankUser"
)

func BankUserContext(ctx context.Context, bankUser model.User) context.Context {
	return context.WithValue(ctx, BankUserKey, &bankUser)
}

func BankUserFromContext(ctx context.Context) model.User {
	switch bankUser := ctx.Value(BankUserKey).(type) {
	case *model.User:
		return *bankUser

	default:
		logger.Logger(ctx).
			Panic("Unable to get BankUser from context")

		return model.User{}
	}
}
