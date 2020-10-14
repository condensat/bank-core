// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"reflect"
	"testing"
	"time"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query/tests"

	"github.com/jinzhu/gorm"
)

func TestAddBatch(t *testing.T) {
	const databaseName = "TestAddBatch"
	t.Parallel()

	db := tests.Setup(databaseName, WithdrawModel())
	defer tests.Teardown(db, databaseName)

	type args struct {
		network model.BatchNetwork
		data    model.BatchData
	}
	tests := []struct {
		name    string
		args    args
		want    model.Batch
		wantErr bool
	}{
		{"default", args{}, createBatch("", ""), true},

		{"default_data", args{model.BatchNetworkBitcoin, ""}, createBatch(model.BatchNetworkBitcoin, "{}"), false},
		{"valid", args{model.BatchNetworkBitcoin, `{"foo": "bar"}`}, createBatch(model.BatchNetworkBitcoin, `{"foo": "bar"}`), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddBatch(db, tt.args.network, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("AddBatch() wrong Timestamp %v", got.Timestamp)
				}
			}

			tt.want.ID = got.ID
			tt.want.Timestamp = got.Timestamp
			tt.want.ExecuteAfter = got.ExecuteAfter
			tt.want.Capacity = got.Capacity
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBatch(t *testing.T) {
	const databaseName = "TestGetBatch"
	t.Parallel()

	db := tests.Setup(databaseName, WithdrawModel())
	defer tests.Teardown(db, databaseName)

	ref, _ := AddBatch(db, "bitcoin", "")

	type args struct {
		ID model.BatchID
	}
	tests := []struct {
		name    string
		args    args
		want    model.Batch
		wantErr bool
	}{
		{"default", args{}, model.Batch{}, true},
		{"ref", args{ref.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBatch(db, tt.args.ID)

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("GetBatch() wrong Timestamp %v", got.Timestamp)
				}

				if got.Capacity != DefaultBatchCapacity {
					t.Errorf("GetBatch() wrong default capacity = %v, want %v", got.Capacity, DefaultBatchCapacity)
				}
				if !got.ExecuteAfter.After(got.Timestamp.Add(DefaultBatchExecutionDelay).Add(-30 * time.Second)) {
					t.Errorf("GetBatch() wrong default execute after = %v", got.ExecuteAfter)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchBatchReady(t *testing.T) {
	const databaseName = "TestFetchBatchReady"
	t.Parallel()

	db := tests.Setup(databaseName, WithdrawModel())
	defer tests.Teardown(db, databaseName)

	// add not ready
	ref, _ := AddBatch(db, "bitcoin", "")
	_, _ = AddBatchInfo(db, ref.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")

	// add ready
	ready, _ := AddBatch(db, "bitcoin", "")
	ready = updateExecuteAfter(db, ready, ready.Timestamp.Add(-time.Minute), ready.Timestamp.Add(-10*time.Second))
	_, _ = AddBatchInfo(db, ready.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")

	tests := []struct {
		name    string
		want    []model.Batch
		wantErr bool
	}{
		{"ready", []model.Batch{ready}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchBatchReady(db)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchBatchReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchBatchReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchBatchByLastStatus(t *testing.T) {
	const databaseName = "TestFetchBatchByLastStatus"
	t.Parallel()

	db := tests.Setup(databaseName, WithdrawModel())
	defer tests.Teardown(db, databaseName)

	// add created
	created, _ := AddBatch(db, "bitcoin", "")
	_, _ = AddBatchInfo(db, created.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")
	// add ready
	ready, _ := AddBatch(db, "bitcoin", "")
	_, _ = AddBatchInfo(db, ready.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")
	_, _ = AddBatchInfo(db, ready.ID, model.BatchStatusReady, model.BatchInfoCrypto, "")
	// add processing
	processing, _ := AddBatch(db, "bitcoin", "")
	_, _ = AddBatchInfo(db, processing.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")
	_, _ = AddBatchInfo(db, processing.ID, model.BatchStatusReady, model.BatchInfoCrypto, "")
	_, _ = AddBatchInfo(db, processing.ID, model.BatchStatusProcessing, model.BatchInfoCrypto, "")

	type args struct {
		status model.BatchStatus
	}
	tests := []struct {
		name    string
		args    args
		want    []model.Batch
		wantErr bool
	}{
		{"created", args{model.BatchStatusCreated}, []model.Batch{created}, false},
		{"ready", args{model.BatchStatusReady}, []model.Batch{ready}, false},
		{"processing", args{model.BatchStatusProcessing}, []model.Batch{processing}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchBatchByLastStatus(db, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchBatchByLastStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchBatchByLastStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestListBatchNetworksByStatus(t *testing.T) {
	const databaseName = "TestListBatchNetworksByStatus"
	t.Parallel()

	db := tests.Setup(databaseName, WithdrawModel())
	defer tests.Teardown(db, databaseName)

	ref1, _ := AddBatch(db, model.BatchNetworkBitcoin, "")
	ref2, _ := AddBatch(db, model.BatchNetworkBitcoinTestnet, "")
	ref3, _ := AddBatch(db, model.BatchNetworkBitcoinLiquid, "")
	ref4, _ := AddBatch(db, model.BatchNetworkBitcoinLightning, "")

	_, _ = AddBatchInfo(db, ref1.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")
	_, _ = AddBatchInfo(db, ref2.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")
	_, _ = AddBatchInfo(db, ref3.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")
	_, _ = AddBatchInfo(db, ref4.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "")

	type args struct {
		status model.BatchStatus
	}
	tests := []struct {
		name    string
		args    args
		want    []model.BatchNetwork
		wantErr bool
	}{
		{"default", args{}, nil, true},

		{"empty", args{model.BatchStatusSettled}, nil, false},
		{"valid", args{model.BatchStatusCreated}, []model.BatchNetwork{ref1.Network, ref2.Network, ref3.Network, ref4.Network}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListBatchNetworksByStatus(db, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBatchNetworksByStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListBatchNetworksByStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createBatch(network model.BatchNetwork, data model.BatchData) model.Batch {
	return model.Batch{
		Network: network,
		Data:    data,
	}
}

func updateExecuteAfter(db database.Context, ready model.Batch, newTimestamp, newExecuteAfter time.Time) model.Batch {
	gdb := db.DB().(*gorm.DB)

	ready.Timestamp = newTimestamp
	ready.ExecuteAfter = newExecuteAfter

	// update
	_ = gdb.Model(&model.Batch{}).
		Where(model.Batch{ID: ready.ID}).
		Update(&ready).Error

	// get from db
	_ = gdb.Model(&model.Batch{}).
		Where(model.Batch{ID: ready.ID}).
		First(&ready).Error

	return ready
}
