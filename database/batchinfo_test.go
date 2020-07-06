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

func TestAddBatchInfo(t *testing.T) {
	const databaseName = "TestAddBatchInfo"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	ref, _ := AddBatch(db, "")

	type args struct {
		batchID model.BatchID
		status  model.BatchStatus
		data    model.BatchInfoData
	}
	tests := []struct {
		name    string
		args    args
		want    model.BatchInfo
		wantErr bool
	}{
		{"default", args{}, model.BatchInfo{}, true},
		{"invalid_status", args{ref.ID, "", "{}"}, model.BatchInfo{}, true},

		{"valid_data", args{ref.ID, model.BatchStatusCreated, ""}, createBatchInfo(ref.ID, model.BatchStatusCreated, "{}"), false},
		{"valid", args{ref.ID, model.BatchStatusCreated, "{}"}, createBatchInfo(ref.ID, model.BatchStatusCreated, "{}"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddBatchInfo(db, tt.args.batchID, tt.args.status, tt.args.data)

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("AddBatchInfo() wrong Timestamp %v", got.Timestamp)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("AddBatchInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.want.ID = got.ID
			tt.want.Timestamp = got.Timestamp
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddBatchInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBatchInfo(t *testing.T) {
	const databaseName = "TestGetBatchInfo"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	batch, _ := AddBatch(db, "{}")

	ref, _ := AddBatchInfo(db, batch.ID, model.BatchStatusCreated, "{}")

	type args struct {
		ID model.BatchInfoID
	}
	tests := []struct {
		name    string
		args    args
		want    model.BatchInfo
		wantErr bool
	}{
		{"default", args{}, model.BatchInfo{}, true},
		{"ref", args{ref.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBatchInfo(db, tt.args.ID)

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("GetBatchInfo() wrong Timestamp %v", got.Timestamp)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatchInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBatchInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBatchHistory(t *testing.T) {
	const databaseName = "TestGetBatchHistory"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	ref, _ := AddBatch(db, "{}")

	ref1, _ := AddBatchInfo(db, ref.ID, model.BatchStatusCreated, "{}")
	ref2, _ := AddBatchInfo(db, ref.ID, model.BatchStatusProcessing, "{}")
	ref3, _ := AddBatchInfo(db, ref.ID, model.BatchStatusCanceled, "{}")
	ref4, _ := AddBatchInfo(db, ref.ID, model.BatchStatusSettled, "{}")

	type args struct {
		batchID model.BatchID
	}
	tests := []struct {
		name    string
		args    args
		want    []model.BatchInfo
		wantErr bool
	}{
		{"default", args{}, nil, true},
		{"ref", args{ref.ID}, createBatchInfoList(ref1, ref2, ref3, ref4), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBatchHistory(db, tt.args.batchID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatchHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBatchHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createBatchInfo(batchID model.BatchID, status model.BatchStatus, data model.BatchInfoData) model.BatchInfo {
	return model.BatchInfo{
		BatchID: batchID,
		Status:  status,
		Data:    data,
	}
}

func createBatchInfoList(list ...model.BatchInfo) []model.BatchInfo {
	return append([]model.BatchInfo{}, list...)
}
