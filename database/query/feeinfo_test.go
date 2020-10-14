// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query/tests"
)

func TestAddOrUpdateFeeInfo(t *testing.T) {
	const databaseName = "TestAddOrUpdateFeeInfo"
	t.Parallel()

	db := tests.Setup(databaseName, FeeModel())
	defer tests.Teardown(db, databaseName)

	ref := model.FeeInfo{
		Currency: "CURR",
		Minimum:  1.0,
		Rate:     model.DefaultFeeRate,
	}

	type args struct {
		feeInfo model.FeeInfo
	}
	tests := []struct {
		name    string
		args    args
		want    model.FeeInfo
		wantErr bool
	}{
		{"default", args{}, model.FeeInfo{}, true},
		{"valid", args{ref}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddOrUpdateFeeInfo(db, tt.args.feeInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrUpdateFeeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddOrUpdateFeeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeInfoExists(t *testing.T) {
	const databaseName = "TestFeeInfoExists"
	t.Parallel()

	db := tests.Setup(databaseName, FeeModel())
	defer tests.Teardown(db, databaseName)

	ref := model.FeeInfo{
		Currency: "CURR",
		Minimum:  1.0,
		Rate:     model.DefaultFeeRate,
	}

	_, _ = AddOrUpdateFeeInfo(db, ref)

	type args struct {
		currency model.CurrencyName
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"default", args{}, false},
		{"notfound", args{"NEW_CURR"}, false},

		{"found", args{ref.Currency}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := FeeInfoExists(db, tt.args.currency); got != tt.want {
				t.Errorf("FeeInfoExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFeeInfo(t *testing.T) {
	const databaseName = "TestGetFeeInfo"
	t.Parallel()

	db := tests.Setup(databaseName, FeeModel())
	defer tests.Teardown(db, databaseName)

	ref := model.FeeInfo{
		Currency: "CURR",
		Minimum:  1.0,
		Rate:     model.DefaultFeeRate,
	}

	_, _ = AddOrUpdateFeeInfo(db, ref)

	type args struct {
		currency model.CurrencyName
	}
	tests := []struct {
		name    string
		args    args
		want    model.FeeInfo
		wantErr bool
	}{
		{"default", args{}, model.FeeInfo{}, true},
		{"notfound", args{"NEW_CURR"}, model.FeeInfo{}, true},

		{"found", args{ref.Currency}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFeeInfo(db, tt.args.currency)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFeeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
