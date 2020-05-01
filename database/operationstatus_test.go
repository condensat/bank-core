// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func TestAddOrUpdateOperationStatus(t *testing.T) {
	const databaseName = "TestAddOrUpdateOperationStatus"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	open, err := AddOrUpdateOperationStatus(db, model.OperationStatus{OperationInfoID: 42, State: "open"})
	if err != nil {
		t.Errorf("Unable to create reference data")
	}

	close := updateOperationSate(open, "close")

	type args struct {
		operation model.OperationStatus
	}
	tests := []struct {
		name    string
		args    args
		want    model.OperationStatus
		wantErr bool
	}{
		{"default", args{}, model.OperationStatus{}, true},
		{"invalidState", args{model.OperationStatus{OperationInfoID: 1, State: ""}}, model.OperationStatus{}, true},

		{"valid", args{model.OperationStatus{OperationInfoID: 42, State: "close"}}, close, false},
		{"update", args{updateOperationSate(open, "close")}, close, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddOrUpdateOperationStatus(db, tt.args.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrUpdateOperationStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.want.LastUpdate = got.LastUpdate
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddOrUpdateOperationStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOperationStatus(t *testing.T) {
	const databaseName = "TestGetOperationStatus"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	const infoID = model.OperationInfoID(42)
	ref1, _ := AddOrUpdateOperationStatus(db, model.OperationStatus{OperationInfoID: infoID, State: "state"})

	type args struct {
		operationInfoID model.OperationInfoID
	}
	tests := []struct {
		name    string
		args    args
		want    model.OperationStatus
		wantErr bool
	}{
		{"default", args{}, model.OperationStatus{}, true},
		{"notExists", args{1337}, model.OperationStatus{}, true},

		{"valid", args{ref1.OperationInfoID}, ref1, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOperationStatus(db, tt.args.operationInfoID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOperationStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindActiveOperationInfo(t *testing.T) {
	const databaseName = "TestFindActiveOperationInfo"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	tests := []struct {
		name    string
		want    []model.OperationInfo
		wantErr bool
	}{
		{"default", nil, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindActiveOperationInfo(db)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindActiveOperationInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindActiveOperationInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindActiveOperationStatus(t *testing.T) {
	const databaseName = "TestFindActiveOperationStatus"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	_, _ = AddOrUpdateOperationStatus(db, model.OperationStatus{OperationInfoID: 42, State: "state", Accounted: "settled"})
	active, _ := AddOrUpdateOperationStatus(db, model.OperationStatus{OperationInfoID: 43, State: "state", Accounted: "active"})

	tests := []struct {
		name    string
		want    []model.OperationStatus
		wantErr bool
	}{
		{"active", []model.OperationStatus{active}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindActiveOperationStatus(db)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindActiveOperationStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindActiveOperationStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func updateOperationSate(operation model.OperationStatus, state string) model.OperationStatus {
	result := operation
	result.State = state
	return result
}
