// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"testing"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"
)

func TestCreateAccount(t *testing.T) {
	const databaseName = "TestCreateAccount"
	t.Parallel()

	ctx := setup(context.Background(), databaseName, AccountModel())
	defer teardown(ctx, databaseName)

	data := createTestAccountData(ctx)

	type args struct {
		ctx     context.Context
		account model.Account
	}
	tests := []struct {
		name    string
		args    args
		validID bool
		wantErr bool
	}{
		{"Default", args{ctx, model.Account{}}, false, true},
		{"Valid", args{ctx, model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name}}, true, false},

		{"SameUser", args{ctx, model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name}}, true, false},
		{"SecondCurr", args{ctx, model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[1].Name}}, true, false},

		{"SecondUser", args{ctx, model.Account{UserID: data.Users[1].ID, CurrencyName: data.Currencies[0].Name}}, true, false},
		{"SecondUserSecondCurr", args{ctx, model.Account{UserID: data.Users[1].ID, CurrencyName: data.Currencies[1].Name}}, true, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateAccount(tt.args.ctx, tt.args.account)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got.ID != 0) != tt.validID {
				t.Errorf("CreateAccount() = %v, unexpected ID", got.ID)
			}
		})
	}
}

type AccountTestData struct {
	Users      []model.User
	Currencies []model.Currency
}

func createTestAccountData(ctx context.Context) AccountTestData {
	var data AccountTestData

	db := appcontext.Database(ctx)
	userTest1, _ := FindOrCreateUser(ctx, db, "test1", "test1@condensat.tech")
	userTest2, _ := FindOrCreateUser(ctx, db, "test2", "test2@condensat.tech")
	currTest1, _ := AddOrUpdateCurrency(ctx, model.NewCurrency("TBTC1", FlagCurencyAvailable))
	currTest2, _ := AddOrUpdateCurrency(ctx, model.NewCurrency("TBTC2", FlagCurencyAvailable))

	data.Users = append(data.Users, *userTest1)
	data.Users = append(data.Users, *userTest2)
	data.Currencies = append(data.Currencies, currTest1)
	data.Currencies = append(data.Currencies, currTest2)

	return data
}
