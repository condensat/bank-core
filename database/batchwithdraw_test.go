// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func TestAddWithdrawToBatch(t *testing.T) {
	const databaseName = "TestAddWithdrawToBatch"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	batch, _ := AddBatch(db, "{}")
	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	w1, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")
	w2, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")
	w3, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")
	w4, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	type args struct {
		batchID   model.BatchID
		withdraws []model.WithdrawID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"default", args{}, true},
		{"ref", args{batch.ID, createWithdrawIDList(w1.ID, w2.ID, w3.ID, w4.ID)}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if err := AddWithdrawToBatch(db, tt.args.batchID, tt.args.withdraws...); (err != nil) != tt.wantErr {
				t.Errorf("AddWithdrawToBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetBatchWithdraws(t *testing.T) {
	const databaseName = "TestGetBatchWithdraws"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	batch, _ := AddBatch(db, "{}")

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	w1, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")
	w2, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")
	w3, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")
	w4, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	withdraws := createWithdrawIDList(w1.ID, w2.ID, w3.ID, w4.ID)

	_ = AddWithdrawToBatch(db, batch.ID, withdraws...)

	type args struct {
		batchID model.BatchID
	}
	tests := []struct {
		name    string
		args    args
		want    []model.WithdrawID
		wantErr bool
	}{
		{"default", args{}, nil, true},
		{"ref", args{batch.ID}, withdraws, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBatchWithdraws(db, tt.args.batchID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatchWithdraws() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBatchWithdraws() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createWithdrawIDList(withdraws ...model.WithdrawID) []model.WithdrawID {
	return append([]model.WithdrawID{}, withdraws...)
}
