// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/condensat/bank-core/database/model"
)

func (p *Database) Migrate(models []model.Model) error {
	var interfaces []interface{}
	for _, model := range models {
		interfaces = append(interfaces, model)
	}
	return p.db.AutoMigrate(
		interfaces...,
	).Error
}

func AccountModel() []model.Model {
	return []model.Model{
		model.Model(new(model.User)),
		model.Model(new(model.Currency)),
		model.Model(new(model.Account)),
	}
}

func AccountStateModel() []model.Model {
	return append(AccountModel(), new(model.AccountState))
}

func AccountOperationModel() []model.Model {
	return append(AccountStateModel(), new(model.AccountOperation))
}

func CurrencyModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Currency)),
		model.Model(new(model.CurrencyRate)),
	}
}

func CryptoAddressModel() []model.Model {
	return []model.Model{
		model.Model(new(model.CryptoAddress)),
	}
}

func OperationInfoModel() []model.Model {
	return append(CryptoAddressModel(), []model.Model{
		model.Model(new(model.OperationInfo)),
		model.Model(new(model.OperationStatus)),
	}...)
}

func AssetModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Asset)),
		model.Model(new(model.AssetInfo)),
		model.Model(new(model.AssetIcon)),
	}
}

func SwapModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Swap)),
		model.Model(new(model.SwapInfo)),
	}
}

func WithdrawModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Withdraw)),
		model.Model(new(model.WithdrawInfo)),
		model.Model(new(model.Fee)),
		model.Model(new(model.Batch)),
		model.Model(new(model.BatchInfo)),
		model.Model(new(model.BatchWithdraw)),
	}
}
