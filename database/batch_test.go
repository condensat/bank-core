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

func TestAddBatch(t *testing.T) {
	const databaseName = "TestAddBatch"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	type args struct {
		data model.BatchData
	}
	tests := []struct {
		name    string
		args    args
		want    model.Batch
		wantErr bool
	}{
		{"default", args{}, createBatch("{}"), false},

		{"default_data", args{""}, createBatch("{}"), false},
		{"valid", args{`{"foo": "bar"}`}, createBatch(`{"foo": "bar"}`), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddBatch(db, tt.args.data)
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBatch(t *testing.T) {
	const databaseName = "TestGetBatch"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	ref, _ := AddBatch(db, "")

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

func createBatch(data model.BatchData) model.Batch {
	return model.Batch{
		Data: data,
	}
}
