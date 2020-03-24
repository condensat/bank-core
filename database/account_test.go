// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"sort"
	"testing"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"
)

func Test_accountColumnNames(t *testing.T) {
	t.Parallel()

	fields := getSortedTypeFileds(reflect.TypeOf(model.Account{}))
	names := accountColumnNames()
	sort.Strings(names)

	if !reflect.DeepEqual(names, fields) {
		t.Errorf("columnsNames() = %v, want %v", names, fields)
	}
}

func TestCreateAccount(t *testing.T) {
	const databaseName = "TestCreateAccount"
	t.Parallel()

	db := setup(databaseName, AccountModel())
	defer teardown(db, databaseName)

	data := createTestAccountData(db)

	type args struct {
		account model.Account
	}
	tests := []struct {
		name    string
		args    args
		validID bool
		wantErr bool
	}{
		{"Default", args{model.Account{}}, false, true},
		{"Valid", args{model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name, Name: data.Names[0]}}, true, false},
		{"Duplicate", args{model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name, Name: data.Names[0]}}, false, true},

		{"SameUser", args{model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name, Name: data.Names[1]}}, true, false},
		{"SecondCurr", args{model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[1].Name, Name: data.Names[0]}}, true, false},

		{"SecondUser", args{model.Account{UserID: data.Users[1].ID, CurrencyName: data.Currencies[0].Name, Name: data.Names[0]}}, true, false},
		{"SecondUserSecondCurr", args{model.Account{UserID: data.Users[1].ID, CurrencyName: data.Currencies[1].Name, Name: data.Names[0]}}, true, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateAccount(db, tt.args.account)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got.ID != 0) != tt.validID {
				t.Errorf("CreateAccount() = %v, unexpected ID", got.ID)
			}
			if !tt.wantErr && len(tt.args.account.Name) == 0 && got.Name != AccountNameDefault {
				t.Errorf("CreateAccount() = %v, unexpected default account name", got.ID)
			}
		})
	}
}

func TestAccountsExists(t *testing.T) {
	const databaseName = "TestAccountsExists"
	t.Parallel()

	db := setup(databaseName, AccountModel())
	defer teardown(db, databaseName)

	data := createTestAccountData(db)

	refAccount := model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name, Name: data.Names[0]}

	_, _ = CreateAccount(db, refAccount)

	type args struct {
		userID   uint64
		currency string
		name     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Default", args{0, "", ""}, false},
		{"Valid", args{refAccount.UserID, refAccount.CurrencyName, refAccount.Name}, true},

		{"InvalidUserID", args{0, refAccount.Name, refAccount.Name}, false},
		{"InvalidCurrency", args{refAccount.UserID, "", refAccount.Name}, false},
		{"InvalidName", args{refAccount.UserID, refAccount.CurrencyName, "not-default"}, false},

		{"InvalidUserIDCurrency", args{0, "", refAccount.Name}, false},
		{"InvalidCurrencyName", args{refAccount.UserID, "", "not-default"}, false},
		{"InvalidUserIDName", args{0, refAccount.CurrencyName, "not-default"}, false},

		{"ValidWildcard", args{refAccount.UserID, refAccount.CurrencyName, AccountNameWildcard}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := AccountsExists(db, tt.args.userID, tt.args.currency, tt.args.name); got != tt.want {
				t.Errorf("AccountsExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryAccountList(t *testing.T) {
	const databaseName = "TestQueryAccountList"
	t.Parallel()

	db := setup(databaseName, AccountModel())
	defer teardown(db, databaseName)

	data := createTestAccountData(db)

	refAccount := model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name, Name: data.Names[1]}
	_, _ = CreateAccount(db, refAccount)

	refAccount = model.Account{UserID: data.Users[0].ID, CurrencyName: data.Currencies[0].Name, Name: data.Names[0]}
	_, _ = CreateAccount(db, refAccount)

	type args struct {
		userID   uint64
		currency string
		name     string
	}
	tests := []struct {
		name    string
		args    args
		count   int
		wantErr bool
	}{
		{"Default", args{0, "", ""}, 0, true},
		{"Valid", args{refAccount.UserID, refAccount.CurrencyName, refAccount.Name}, 1, false},

		{"InvalidUserID", args{0, refAccount.Name, refAccount.Name}, 0, true},
		{"InvalidCurrency", args{refAccount.UserID, "", refAccount.Name}, 0, false},
		{"InvalidName", args{refAccount.UserID, refAccount.CurrencyName, "not-default"}, 0, false},

		{"ValidWildcard", args{refAccount.UserID, refAccount.CurrencyName, AccountNameWildcard}, 2, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := QueryAccountList(db, tt.args.userID, tt.args.currency, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryAccountList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.count {
				t.Errorf("QueryAccountList() = %v, want %v", len(got), tt.count)
			}
		})
	}
}

type AccountTestData struct {
	Users      []model.User
	Currencies []model.Currency
	Names      []string
}

func createTestAccountData(db bank.Database) AccountTestData {
	var data AccountTestData

	userTest1, _ := FindOrCreateUser(db, "test1", "test1@condensat.tech")
	userTest2, _ := FindOrCreateUser(db, "test2", "test2@condensat.tech")
	currTest1, _ := AddOrUpdateCurrency(db, model.NewCurrency("TBTC1", FlagCurencyAvailable))
	currTest2, _ := AddOrUpdateCurrency(db, model.NewCurrency("TBTC2", FlagCurencyAvailable))

	data.Users = append(data.Users, *userTest1)
	data.Users = append(data.Users, *userTest2)
	data.Currencies = append(data.Currencies, currTest1)
	data.Currencies = append(data.Currencies, currTest2)
	data.Names = append(data.Names, "") // empty account name is "default"
	data.Names = append(data.Names, "Vault")

	return data
}
