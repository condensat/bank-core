// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"
)

func TestAddOrUpdateAccountState(t *testing.T) {
	const databaseName = "TestAddOrUpdateAccountState"
	t.Parallel()

	db := setup(databaseName, AccountOperationModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)

	accountToUpdate, err := AddOrUpdateAccountState(db, createAccountState(data.Accounts[1].ID, model.AccountStatusDisabled))
	if err != nil {
		t.Errorf("AddOrUpdateAccountState() error = %v", err)
	}

	type args struct {
		accountSate model.AccountState
	}
	tests := []struct {
		name    string
		args    args
		want    model.AccountState
		wantErr bool
	}{
		// model.AccountState{AccountID: data.Accounts[0].ID, State: model.AccountStatusNormal}
		{"Default", args{model.AccountState{}}, model.AccountState{}, true},
		{"InvlalidAccountID", args{createAccountState(0, model.AccountStatusInvalid)}, model.AccountState{}, true},
		{"InvlalidStatus", args{createAccountState(data.Accounts[0].ID, model.AccountStatusInvalid)}, model.AccountState{}, true},

		{"Created", args{createAccountState(data.Accounts[0].ID, model.AccountStatusCreated)}, createAccountState(data.Accounts[0].ID, model.AccountStatusCreated), false},
		{"Normal", args{createAccountState(data.Accounts[0].ID, model.AccountStatusNormal)}, createAccountState(data.Accounts[0].ID, model.AccountStatusNormal), false},
		{"Locked", args{createAccountState(data.Accounts[0].ID, model.AccountStatusLocked)}, createAccountState(data.Accounts[0].ID, model.AccountStatusLocked), false},
		{"Disabled", args{createAccountState(data.Accounts[0].ID, model.AccountStatusDisabled)}, createAccountState(data.Accounts[0].ID, model.AccountStatusDisabled), false},

		{"Update", args{updateAccountState(accountToUpdate, model.AccountStatusNormal)}, createAccountState(accountToUpdate.AccountID, model.AccountStatusNormal), false},
		{"InvalidUpdate", args{updateAccountState(accountToUpdate, model.AccountStatusInvalid)}, model.AccountState{}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddOrUpdateAccountState(db, tt.args.accountSate)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrUpdateAccountState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddOrUpdateAccountState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAccountStatusByAccountID(t *testing.T) {
	const databaseName = "TestGetAccountStatusByAccountID"
	t.Parallel()

	db := setup(databaseName, AccountOperationModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)

	refAccountState := model.AccountState{AccountID: data.Accounts[0].ID, State: model.AccountStatusNormal}

	_, _ = AddOrUpdateAccountState(db, refAccountState)

	type args struct {
		accountID model.AccountID
	}
	tests := []struct {
		name    string
		args    args
		want    model.AccountState
		wantErr bool
	}{
		{"Default", args{}, model.AccountState{}, true},
		{"Valid", args{refAccountState.AccountID}, refAccountState, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAccountStatusByAccountID(db, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountStatusByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountStatusByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

type AccountStateTestData struct {
	Users      []model.User
	Currencies []model.Currency
	Accounts   []model.Account
}

func createAccountState(accountID model.AccountID, state model.AccountStatus) model.AccountState {
	return model.AccountState{
		AccountID: accountID,
		State:     state,
	}
}

func updateAccountState(accountState model.AccountState, newState model.AccountStatus) model.AccountState {
	return model.AccountState{
		AccountID: accountState.AccountID,
		State:     newState,
	}
}

func createTestAccountStateData(db bank.Database) AccountStateTestData {
	var data AccountStateTestData

	userTest1, _ := FindOrCreateUser(db, model.User{Name: "test1", Email: "test1@condensat.tech"})
	userTest2, _ := FindOrCreateUser(db, model.User{Name: "test2", Email: "test2@condensat.tech"})
	currTest1, _ := AddOrUpdateCurrency(db, model.NewCurrency("TBTC1", FlagCurencyAvailable, 1, 2))
	currTest2, _ := AddOrUpdateCurrency(db, model.NewCurrency("TBTC2", FlagCurencyAvailable, 1, 2))
	accountTest1, _ := CreateAccount(db, model.Account{UserID: userTest1.ID, CurrencyName: currTest1.Name, Name: "accountTest1"})
	accountTest2, _ := CreateAccount(db, model.Account{UserID: userTest1.ID, CurrencyName: currTest2.Name, Name: "accountTest2"})
	accountTest3, _ := CreateAccount(db, model.Account{UserID: userTest2.ID, CurrencyName: currTest1.Name, Name: "accountTest3"})
	accountTest4, _ := CreateAccount(db, model.Account{UserID: userTest2.ID, CurrencyName: currTest2.Name, Name: "accountTest4"})

	data.Users = append(data.Users, userTest1)
	data.Users = append(data.Users, userTest2)
	data.Currencies = append(data.Currencies, currTest1)
	data.Currencies = append(data.Currencies, currTest2)
	data.Accounts = append(data.Accounts, accountTest1)
	data.Accounts = append(data.Accounts, accountTest2)
	data.Accounts = append(data.Accounts, accountTest3)
	data.Accounts = append(data.Accounts, accountTest4)

	return data
}
