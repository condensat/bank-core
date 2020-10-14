// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
)

func UserModel() []database.Model {
	return []database.Model{
		database.Model(new(model.User)),
		database.Model(new(model.UserRole)),
	}
}

func AccountModel() []database.Model {
	return append(UserModel(), []database.Model{
		database.Model(new(model.Currency)),
		database.Model(new(model.Account)),
	}...)
}

func AccountStateModel() []database.Model {
	return append(AccountModel(), new(model.AccountState))
}

func AccountOperationModel() []database.Model {
	return append(AccountStateModel(), new(model.AccountOperation))
}

func CurrencyModel() []database.Model {
	return []database.Model{
		database.Model(new(model.Currency)),
		database.Model(new(model.CurrencyRate)),
	}
}

func CryptoAddressModel() []database.Model {
	return []database.Model{
		database.Model(new(model.CryptoAddress)),
	}
}

func SsmAddressModel() []database.Model {
	return []database.Model{
		database.Model(new(model.SsmAddress)),
		database.Model(new(model.SsmAddressInfo)),
		database.Model(new(model.SsmAddressState)),
	}
}

func OperationInfoModel() []database.Model {
	return append(CryptoAddressModel(), []database.Model{
		database.Model(new(model.OperationInfo)),
		database.Model(new(model.OperationStatus)),
	}...)
}

func AssetModel() []database.Model {
	return []database.Model{
		database.Model(new(model.Asset)),
		database.Model(new(model.AssetInfo)),
		database.Model(new(model.AssetIcon)),
	}
}

func SwapModel() []database.Model {
	return []database.Model{
		database.Model(new(model.Swap)),
		database.Model(new(model.SwapInfo)),
	}
}

func WithdrawModel() []database.Model {
	return append(AccountOperationModel(), []database.Model{
		database.Model(new(model.Withdraw)),
		database.Model(new(model.WithdrawInfo)),
		database.Model(new(model.WithdrawTarget)),
		database.Model(new(model.Fee)),
		database.Model(new(model.FeeInfo)),
		database.Model(new(model.Batch)),
		database.Model(new(model.BatchInfo)),
		database.Model(new(model.BatchWithdraw)),
	}...)
}

func FeeModel() []database.Model {
	return []database.Model{
		database.Model(new(model.Fee)),
		database.Model(new(model.FeeInfo)),
	}
}
