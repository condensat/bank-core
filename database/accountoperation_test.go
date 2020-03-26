// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"
)

func TestAppendAccountOperation(t *testing.T) {
	const databaseName = "TestAppendAccountOperation"
	t.Parallel()

	db := setup(databaseName, AccountOperationModel())
	defer teardown(db, databaseName)

	data := createTestAccountOperationData(db)
	refAccountOperation := createOperation(data.Accounts[0].ID, 0, 1.0, 1.0)

	refInvalidAccountID := cloneOperation(refAccountOperation)
	refInvalidAccountID.AccountID = 0

	refNotExistingAccountID := cloneOperation(refAccountOperation)
	refNotExistingAccountID.AccountID = 42

	refInvalidPreCheck := cloneOperation(refAccountOperation)
	*refInvalidPreCheck.Balance = 0.0

	type args struct {
		db        bank.Database
		operation model.AccountOperation
	}
	tests := []struct {
		name    string
		args    args
		want    model.AccountOperation
		wantErr bool
	}{
		{"Default", args{}, model.AccountOperation{}, true},
		{"NilDB", args{nil, refAccountOperation}, model.AccountOperation{}, true},
		{"InvalidAccountID", args{db, refInvalidAccountID}, model.AccountOperation{}, true},
		{"NotExistingAccountID", args{db, refNotExistingAccountID}, model.AccountOperation{}, true},
		{"InvalidPreCheck", args{db, refInvalidPreCheck}, model.AccountOperation{}, true},

		{"Valid", args{db, refAccountOperation}, refAccountOperation, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendAccountOperation(tt.args.db, tt.args.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppendAccountOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.want.ID = got.ID
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendAccountOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLastAccountOperation(t *testing.T) {
	const databaseName = "TestGetLastAccountOperation"
	t.Parallel()

	db := setup(databaseName, AccountOperationModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	var ops []model.AccountOperation
	for i := 0; i < len(data.Accounts); i++ {
		ops = append(ops, lastLinkedOperation(createLinkedOperations(db, data.Accounts[i].ID, i+1, 1.0)))
	}

	type args struct {
		db        bank.Database
		accountID model.AccountID
	}
	tests := []struct {
		name       string
		args       args
		wantPrevID model.AccountOperationID
		wantErr    bool
	}{
		{"Default", args{}, 0, true},
		{"NilDB", args{nil, ops[0].AccountID}, 0, true},
		{"InvalidAccountID", args{db, 0}, 0, true},

		{"op1", args{db, ops[0].AccountID}, ops[0].PrevID, false},
		{"op2", args{db, ops[1].AccountID}, ops[1].PrevID, false},
		{"op3", args{db, ops[2].AccountID}, ops[2].PrevID, false},
		{"op4", args{db, ops[3].AccountID}, ops[3].PrevID, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLastAccountOperation(tt.args.db, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastAccountOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.PrevID != tt.wantPrevID {
				t.Errorf("GetLastAccountOperation() PrevID = %v, wantPrevID %v", got.PrevID, tt.wantPrevID)
			}
		})
	}
}

func TestGeAccountHistory(t *testing.T) {
	const databaseName = "TestGeAccountHistory"
	t.Parallel()

	db := setup(databaseName, AccountOperationModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	var ops [][]model.AccountOperation
	for i := 0; i < len(data.Accounts); i++ {
		ops = append(ops, createLinkedOperations(db, data.Accounts[i].ID, i+1, 1.0))
	}

	type args struct {
		db        bank.Database
		accountID model.AccountID
	}
	tests := []struct {
		name    string
		args    args
		want    []model.AccountOperation
		wantErr bool
	}{
		{"Default", args{}, nil, true},
		{"NilDB", args{nil, lastLinkedOperation(ops[0]).AccountID}, nil, true},
		{"InvalidAccountID", args{db, 0}, nil, true},

		{"op1", args{db, lastLinkedOperation(ops[0]).AccountID}, ops[0], false},
		{"op2", args{db, lastLinkedOperation(ops[1]).AccountID}, ops[1], false},
		{"op3", args{db, lastLinkedOperation(ops[2]).AccountID}, ops[2], false},
		{"op4", args{db, lastLinkedOperation(ops[3]).AccountID}, ops[3], false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeAccountHistory(tt.args.db, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeAccountHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GeAccountHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeAccountHistoryRange(t *testing.T) {
	const databaseName = "TestGeAccountHistoryRange"
	t.Parallel()

	db := setup(databaseName, AccountOperationModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	var ops [][]model.AccountOperation
	for i := 0; i < len(data.Accounts); i++ {
		ops = append(ops, createLinkedOperations(db, data.Accounts[i].ID, i+1, 1.0))
	}

	to := time.Now()
	from := to.Add(-10 * time.Second)

	afterTo := to.Add(time.Minute)
	afterFrom := from.Add(time.Minute)

	beforeTo := to.Add(-time.Minute)
	beforeFrom := from.Add(-time.Minute)

	type args struct {
		db        bank.Database
		accountID model.AccountID
		from      time.Time
		to        time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    []model.AccountOperation
		wantErr bool
	}{
		{"Default", args{}, nil, true},
		{"NilDB", args{nil, lastLinkedOperation(ops[0]).AccountID, time.Time{}, time.Time{}}, nil, true},
		{"InvalidAccountID", args{db, 0, from, to}, nil, true},

		{"DefaultRangeOp1", args{db, lastLinkedOperation(ops[0]).AccountID, time.Time{}, time.Time{}}, nil, false},
		{"DefaultRangeOp2", args{db, lastLinkedOperation(ops[1]).AccountID, time.Time{}, time.Time{}}, nil, false},
		{"DefaultRangeOp3", args{db, lastLinkedOperation(ops[2]).AccountID, time.Time{}, time.Time{}}, nil, false},
		{"DefaultRangeOp4", args{db, lastLinkedOperation(ops[3]).AccountID, time.Time{}, time.Time{}}, nil, false},

		{"Rangeop1", args{db, lastLinkedOperation(ops[0]).AccountID, from, to}, ops[0], false},
		{"Rangeop2", args{db, lastLinkedOperation(ops[1]).AccountID, from, to}, ops[1], false},
		{"Rangeop3", args{db, lastLinkedOperation(ops[2]).AccountID, from, to}, ops[2], false},
		{"Rangeop4", args{db, lastLinkedOperation(ops[3]).AccountID, from, to}, ops[3], false},

		{"InvertRangeOp1", args{db, lastLinkedOperation(ops[0]).AccountID, to, from}, ops[0], false},
		{"InvertRangeOp2", args{db, lastLinkedOperation(ops[1]).AccountID, to, from}, ops[1], false},
		{"InvertRangeOp3", args{db, lastLinkedOperation(ops[2]).AccountID, to, from}, ops[2], false},
		{"InvertRangeOp4", args{db, lastLinkedOperation(ops[3]).AccountID, to, from}, ops[3], false},

		{"BeforeRangeOp1", args{db, lastLinkedOperation(ops[0]).AccountID, beforeFrom, beforeTo}, nil, false},
		{"BeforeRangeOp2", args{db, lastLinkedOperation(ops[1]).AccountID, beforeFrom, beforeTo}, nil, false},
		{"BeforeRangeOp3", args{db, lastLinkedOperation(ops[2]).AccountID, beforeFrom, beforeTo}, nil, false},
		{"BeforeRangeOp4", args{db, lastLinkedOperation(ops[3]).AccountID, beforeFrom, beforeTo}, nil, false},

		{"AfterRangeOp1", args{db, lastLinkedOperation(ops[0]).AccountID, afterFrom, afterTo}, nil, false},
		{"AfterRangeOp2", args{db, lastLinkedOperation(ops[1]).AccountID, afterFrom, afterTo}, nil, false},
		{"AfterRangeOp3", args{db, lastLinkedOperation(ops[2]).AccountID, afterFrom, afterTo}, nil, false},
		{"AfterRangeOp4", args{db, lastLinkedOperation(ops[3]).AccountID, afterFrom, afterTo}, nil, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeAccountHistoryRange(tt.args.db, tt.args.accountID, tt.args.from, tt.args.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeAccountHistoryRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GeAccountHistoryRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createOperation(account model.AccountID, prevID model.AccountOperationID, amount, balance model.Float) model.AccountOperation {
	return model.NewAccountOperation(0, prevID, account, model.SynchroneousTypeSync, model.OperationTypeDeposit, 0, time.Now(), amount, balance, 0.0, 0.0)
}

func cloneOperation(operation model.AccountOperation) model.AccountOperation {
	return createOperation(operation.AccountID, operation.PrevID, *operation.Amount, *operation.Balance)
}

func createLinkedOperations(db bank.Database, account model.AccountID, count int, amount model.Float) []model.AccountOperation {
	var list []model.AccountOperation

	var balance model.Float
	for i := 0; i < count; i++ {
		balance += amount
		last := storeOperation(db, createOperation(account, 0, amount, balance))
		if !last.IsValid() {
			panic("Invalid AccountOperation")
		}
		list = append(list, last)
	}

	return list
}

func lastLinkedOperation(list []model.AccountOperation) model.AccountOperation {
	if len(list) == 0 {
		panic("empty list")
	}

	return list[len(list)-1]
}

func storeOperation(db bank.Database, operation model.AccountOperation) model.AccountOperation {
	gdb := getGormDB(db)
	if gdb == nil {
		return model.AccountOperation{}
	}

	err := gdb.Create(&operation).Error
	if err != nil {
		return model.AccountOperation{}
	}

	return operation
}

type AccountOperationTestData struct {
	AccountStateTestData
	AccountStates []model.AccountState
}

func createTestAccountOperationData(db bank.Database) AccountOperationTestData {
	var data AccountOperationTestData
	data.AccountStateTestData = createTestAccountStateData(db)

	accountState1, _ := AddOrUpdateAccountState(db, model.AccountState{AccountID: data.Accounts[0].ID, State: model.AccountStatusNormal})
	accountState2, _ := AddOrUpdateAccountState(db, model.AccountState{AccountID: data.Accounts[1].ID, State: model.AccountStatusNormal})
	accountState3, _ := AddOrUpdateAccountState(db, model.AccountState{AccountID: data.Accounts[2].ID, State: model.AccountStatusNormal})
	accountState4, _ := AddOrUpdateAccountState(db, model.AccountState{AccountID: data.Accounts[3].ID, State: model.AccountStatusNormal})

	data.AccountStates = append(data.AccountStates, accountState1)
	data.AccountStates = append(data.AccountStates, accountState2)
	data.AccountStates = append(data.AccountStates, accountState3)
	data.AccountStates = append(data.AccountStates, accountState4)

	return data
}
