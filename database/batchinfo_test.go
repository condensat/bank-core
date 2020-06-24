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

	ref, _ := AddBatch(db, model.BatchNetworkBitcoin, "")

	type args struct {
		batchID  model.BatchID
		status   model.BatchStatus
		dataType model.DataType
		data     model.BatchInfoData
	}
	tests := []struct {
		name    string
		args    args
		want    model.BatchInfo
		wantErr bool
	}{
		{"default", args{}, model.BatchInfo{}, true},
		{"invalid_status", args{ref.ID, "", model.BatchInfoCrypto, "{}"}, model.BatchInfo{}, true},
		{"invalid_datatype", args{ref.ID, model.BatchStatusCreated, "", "{}"}, model.BatchInfo{}, true},

		{"valid_data", args{ref.ID, model.BatchStatusCreated, model.BatchInfoCrypto, ""}, createBatchInfo(ref.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}"), false},
		{"valid", args{ref.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}"}, createBatchInfo(ref.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddBatchInfo(db, tt.args.batchID, tt.args.status, tt.args.dataType, tt.args.data)

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

	batch, _ := AddBatch(db, model.BatchNetworkBitcoin, "{}")

	ref, _ := AddBatchInfo(db, batch.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}")

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

	ref, _ := AddBatch(db, model.BatchNetworkBitcoin, "{}")

	ref1, _ := AddBatchInfo(db, ref.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}")
	ref2, _ := AddBatchInfo(db, ref.ID, model.BatchStatusProcessing, model.BatchInfoCrypto, "{}")
	ref3, _ := AddBatchInfo(db, ref.ID, model.BatchStatusCanceled, model.BatchInfoCrypto, "{}")
	ref4, _ := AddBatchInfo(db, ref.ID, model.BatchStatusSettled, model.BatchInfoCrypto, "{}")

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

func TestGetBatchInfoByStatusAndType(t *testing.T) {
	const databaseName = "TestGetBatchInfoByStatusAndType"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	ref, _ := AddBatch(db, model.BatchNetworkBitcoin, "{}")

	ref1, _ := AddBatchInfo(db, ref.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}")
	ref2, _ := AddBatchInfo(db, ref.ID, model.BatchStatusProcessing, model.BatchInfoCrypto, "{}")
	ref3, _ := AddBatchInfo(db, ref.ID, model.BatchStatusCanceled, model.BatchInfoCrypto, "{}")
	ref4, _ := AddBatchInfo(db, ref.ID, model.BatchStatusSettled, model.BatchInfoCrypto, "{}")

	type args struct {
		status   model.BatchStatus
		dataType model.DataType
	}
	tests := []struct {
		name    string
		args    args
		want    []model.BatchInfo
		wantErr bool
	}{
		{"default", args{}, nil, true},

		{"invalid_status", args{"", model.BatchInfoCrypto}, nil, true},
		{"invalid_datatype", args{model.BatchStatusCreated, ""}, nil, true},

		{"created", args{"other", model.BatchInfoCrypto}, nil, false},

		{"created", args{model.BatchStatusCreated, model.BatchInfoCrypto}, createBatchInfoList(ref1), false},
		{"processing", args{model.BatchStatusProcessing, model.BatchInfoCrypto}, createBatchInfoList(ref2), false},
		{"canceled", args{model.BatchStatusCanceled, model.BatchInfoCrypto}, createBatchInfoList(ref3), false},
		{"settled", args{model.BatchStatusSettled, model.BatchInfoCrypto}, createBatchInfoList(ref4), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBatchInfoByStatusAndType(db, tt.args.status, tt.args.dataType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatchInfoByStatusAndType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBatchInfoByStatusAndType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLastBatchInfoByStatusAndNetwork(t *testing.T) {
	const databaseName = "TestGetLastBatchInfoByStatusAndNetwork"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	b1, _ := AddBatch(db, model.BatchNetworkBitcoin, "{}")
	b2, _ := AddBatch(db, model.BatchNetworkBitcoin, "{}")

	data, _ := model.EncodeData(&model.BatchInfoCryptoData{
		TxID: "",
	})

	ref1, _ := AddBatchInfo(db, b1.ID, model.BatchStatusCreated, model.BatchInfoCrypto, model.BatchInfoData(data))
	_, _ = AddBatchInfo(db, b2.ID, model.BatchStatusCreated, model.BatchInfoCrypto, model.BatchInfoData(data))
	ref2, _ := AddBatchInfo(db, b2.ID, model.BatchStatusProcessing, model.BatchInfoCrypto, model.BatchInfoData(data))

	type args struct {
		status   model.BatchStatus
		network  model.BatchNetwork
		dataType model.DataType
	}
	tests := []struct {
		name    string
		args    args
		want    []model.BatchInfo
		wantErr bool
	}{
		{"default", args{}, nil, true},

		{"absent", args{model.BatchStatusCreated, model.BatchNetworkSepa, "absent"}, nil, false},
		{"created", args{model.BatchStatusCreated, model.BatchNetworkBitcoin, model.BatchInfoCrypto}, createBatchInfoList(ref1), false},
		{"processing", args{model.BatchStatusProcessing, model.BatchNetworkBitcoin, model.BatchInfoCrypto}, createBatchInfoList(ref2), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLastBatchInfoByStatusAndNetwork(db, tt.args.status, tt.args.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastBatchInfoByStatusAndNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLastBatchInfoByStatusAndNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createBatchInfoList(list ...model.BatchInfo) []model.BatchInfo {
	return append([]model.BatchInfo{}, list...)
}

func createBatchInfo(batchID model.BatchID, status model.BatchStatus, dataType model.DataType, data model.BatchInfoData) model.BatchInfo {
	return model.BatchInfo{
		BatchID: batchID,
		Status:  status,
		Type:    dataType,
		Data:    data,
	}
}
