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

func updateOperationSate(operation model.OperationStatus, state string) model.OperationStatus {
	result := operation
	result.State = state
	return result
}
