// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func TestAddFee(t *testing.T) {
	const databaseName = "TestAddFee"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	ref1, _ := AddWithdraw(db, 42, 1337, 0.1, model.BatchModeNormal, "{}")
	ref2, _ := AddWithdraw(db, 42, 1337, 0.1, model.BatchModeNormal, "{}")

	type args struct {
		withdrawID model.WithdrawID
		amount     model.Float
		data       model.FeeData
	}
	tests := []struct {
		name    string
		args    args
		want    model.Fee
		wantErr bool
	}{
		{"default", args{}, model.Fee{}, true},
		{"invalid_withdraw", args{0, 0.1, ""}, model.Fee{}, true},
		{"invalid_fee", args{ref1.ID, 0.0, ""}, model.Fee{}, true},
		{"negative_fee", args{ref1.ID, -0.1, ""}, model.Fee{}, true},

		{"valid_data", args{ref1.ID, 0.1, ""}, createFee(ref1.ID, 0.1, "{}"), false},
		{"valid", args{ref2.ID, 0.1, "{}"}, createFee(ref2.ID, 0.1, "{}"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddFee(db, tt.args.withdrawID, tt.args.amount, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.want.ID = got.ID
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFee(t *testing.T) {
	const databaseName = "TestGetFee"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	withdraw, _ := AddWithdraw(db, 42, 1337, 0.1, model.BatchModeNormal, "{}")

	ref, _ := AddFee(db, withdraw.ID, 0.1, "{}")

	type args struct {
		ID model.FeeID
	}
	tests := []struct {
		name    string
		args    args
		want    model.Fee
		wantErr bool
	}{
		{"default", args{}, model.Fee{}, true},
		{"ref", args{ref.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFee(db, tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFeeByWithdrawID(t *testing.T) {
	const databaseName = "TestGetFeeByWithdrawID"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	withdraw, _ := AddWithdraw(db, 42, 1337, 0.1, model.BatchModeNormal, "{}")

	ref, _ := AddFee(db, withdraw.ID, 0.1, "{}")

	type args struct {
		withdrawID model.WithdrawID
	}
	tests := []struct {
		name    string
		args    args
		want    model.Fee
		wantErr bool
	}{
		{"default", args{}, model.Fee{}, true},
		{"ref", args{withdraw.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFeeByWithdrawID(db, tt.args.withdrawID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFeeByWithdrawID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeeByWithdrawID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createFee(withdrawID model.WithdrawID, amount model.Float, data model.FeeData) model.Fee {
	return model.Fee{
		WithdrawID: withdrawID,
		Amount:     &amount,
		Data:       data,
	}
}
