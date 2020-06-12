// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"
	"time"

	"github.com/condensat/bank-core/database/model"
)

func TestAddWithdrawInfo(t *testing.T) {
	const databaseName = "TestAddWithdrawInfo"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	ref, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	type args struct {
		withdrawID model.WithdrawID
		status     model.WithdrawStatus
		data       model.WithdrawInfoData
	}
	tests := []struct {
		name    string
		args    args
		want    model.WithdrawInfo
		wantErr bool
	}{
		{"default", args{}, model.WithdrawInfo{}, true},
		{"invalid_withdraw", args{0, model.WithdrawStatusCreated, "{}"}, model.WithdrawInfo{}, true},
		{"invalid_status", args{ref.ID, "", "{}"}, model.WithdrawInfo{}, true},

		{"default_data", args{ref.ID, model.WithdrawStatusCreated, ""}, createWithdrawInfo(ref.ID, model.WithdrawStatusCreated, "{}"), false},
		{"valid", args{ref.ID, model.WithdrawStatusCreated, "{}"}, createWithdrawInfo(ref.ID, model.WithdrawStatusCreated, "{}"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddWithdrawInfo(db, tt.args.withdrawID, tt.args.status, tt.args.data)

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("AddWithdrawInfo() wrong Timestamp %v", got.Timestamp)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("AddWithdrawInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.want.ID = got.ID
			tt.want.Timestamp = got.Timestamp
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddWithdrawInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithdrawInfo(t *testing.T) {
	const databaseName = "TestGetWithdrawInfo"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	withdraw, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	ref, _ := AddWithdrawInfo(db, withdraw.ID, model.WithdrawStatusCreated, "{}")

	type args struct {
		ID model.WithdrawInfoID
	}
	tests := []struct {
		name    string
		args    args
		want    model.WithdrawInfo
		wantErr bool
	}{
		{"default", args{}, model.WithdrawInfo{}, true},
		{"ref", args{ref.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWithdrawInfo(db, tt.args.ID)

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("GetWithdrawInfo() wrong Timestamp %v", got.Timestamp)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithdrawInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithdrawHistory(t *testing.T) {
	const databaseName = "TestGetWithdrawHistory"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	ref, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	ref1, _ := AddWithdrawInfo(db, ref.ID, model.WithdrawStatusCreated, "{}")
	ref2, _ := AddWithdrawInfo(db, ref.ID, model.WithdrawStatusProcessing, "{}")
	ref3, _ := AddWithdrawInfo(db, ref.ID, model.WithdrawStatusCanceled, "{}")
	ref4, _ := AddWithdrawInfo(db, ref.ID, model.WithdrawStatusSettled, "{}")

	type args struct {
		withdrawID model.WithdrawID
	}
	tests := []struct {
		name    string
		args    args
		want    []model.WithdrawInfo
		wantErr bool
	}{
		{"default", args{}, nil, true},
		{"ref", args{ref.ID}, createWithdrawInfoList(ref1, ref2, ref3, ref4), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWithdrawHistory(db, tt.args.withdrawID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithdrawHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createWithdrawInfo(withdrawID model.WithdrawID, status model.WithdrawStatus, data model.WithdrawInfoData) model.WithdrawInfo {
	return model.WithdrawInfo{
		WithdrawID: withdrawID,
		Status:     status,
		Data:       data,
	}
}

func createWithdrawInfoList(list ...model.WithdrawInfo) []model.WithdrawInfo {
	return append([]model.WithdrawInfo{}, list...)
}
