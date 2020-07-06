// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func TestAddWithdrawTarget(t *testing.T) {
	const databaseName = "TestAddWithdrawTarget"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	ref, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	type args struct {
		withdrawID model.WithdrawID
		dataType   model.WithdrawTargetType
		data       model.WithdrawTargetData
	}
	tests := []struct {
		name    string
		args    args
		want    model.WithdrawTarget
		wantErr bool
	}{
		{"default", args{}, model.WithdrawTarget{}, true},
		{"invalid_type", args{ref.ID, "", "{}"}, model.WithdrawTarget{}, true},

		{"valid_data", args{ref.ID, model.WithdrawTargetOnChain, ""}, createWithdrawTarget(ref.ID, model.WithdrawTargetOnChain, "{}"), false},
		{"valid", args{ref.ID, model.WithdrawTargetOnChain, "{}"}, createWithdrawTarget(ref.ID, model.WithdrawTargetOnChain, "{}"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddWithdrawTarget(db, tt.args.withdrawID, tt.args.dataType, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddWithdrawTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.want.ID = got.ID
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddWithdrawTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithdrawTarget(t *testing.T) {
	const databaseName = "TestGetWithdrawTarget"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	withdraw, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	ref, _ := AddWithdrawTarget(db, withdraw.ID, model.WithdrawTargetOnChain, "{}")

	type args struct {
		ID model.WithdrawTargetID
	}
	tests := []struct {
		name    string
		args    args
		want    model.WithdrawTarget
		wantErr bool
	}{
		{"default", args{}, model.WithdrawTarget{}, true},
		{"ref", args{ref.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWithdrawTarget(db, tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithdrawTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithdrawTargetByWithdrawID(t *testing.T) {
	const databaseName = "TestGetWithdrawTargetByWithdrawID"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	withdraw, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	ref, _ := AddWithdrawTarget(db, withdraw.ID, model.WithdrawTargetOnChain, "{}")

	type args struct {
		withdrawID model.WithdrawID
	}
	tests := []struct {
		name    string
		args    args
		want    model.WithdrawTarget
		wantErr bool
	}{
		{"default", args{}, model.WithdrawTarget{}, true},
		{"ref", args{withdraw.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWithdrawTargetByWithdrawID(db, tt.args.withdrawID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawTargetByWithdrawID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithdrawTargetByWithdrawID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createWithdrawTarget(withdrawID model.WithdrawID, dataType model.WithdrawTargetType, data model.WithdrawTargetData) model.WithdrawTarget {
	return model.WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       dataType,
		Data:       data,
	}
}
