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

func TestAddWithdraw(t *testing.T) {
	const databaseName = "TestAddWithdraw"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]
	a3 := data.Accounts[1]

	type args struct {
		from   model.AccountID
		to     model.AccountID
		amount model.Float
		batch  model.BatchMode
		data   model.WithdrawData
	}
	tests := []struct {
		name    string
		args    args
		want    model.Withdraw
		wantErr bool
	}{
		{"default", args{}, model.Withdraw{}, true},
		{"invalid_from", args{0, a2.ID, 0.1, model.BatchModeNormal, "{}"}, model.Withdraw{}, true},
		{"invalid_to", args{a1.ID, 0, 0.1, model.BatchModeNormal, "{}"}, model.Withdraw{}, true},
		{"same_from_to", args{a1.ID, a1.ID, 0.1, model.BatchModeNormal, "{}"}, model.Withdraw{}, true},
		{"wrong_currency", args{a1.ID, a3.ID, 0.1, model.BatchModeNormal, "{}"}, model.Withdraw{}, true},
		{"invalid_amount", args{a1.ID, a2.ID, 0.0, model.BatchModeNormal, "{}"}, model.Withdraw{}, true},
		{"negative_amount", args{a1.ID, a2.ID, -0.1, model.BatchModeNormal, "{}"}, model.Withdraw{}, true},
		{"invalid_batch", args{a1.ID, a2.ID, 0.0, "", "{}"}, model.Withdraw{}, true},

		{"default_data", args{a1.ID, a2.ID, 0.1, model.BatchModeNormal, ""}, createWithdraw(a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}"), false},
		{"valid", args{a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}"}, createWithdraw(a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddWithdraw(db, tt.args.from, tt.args.to, tt.args.amount, tt.args.batch, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddWithdraw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("AddWithdraw() wrong Timestamp %v", got.Timestamp)
				}
			}

			tt.want.ID = got.ID
			tt.want.Timestamp = got.Timestamp
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddWithdraw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithdraw(t *testing.T) {
	const databaseName = "TestGetWithdraw"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	ref, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")

	type args struct {
		ID model.WithdrawID
	}
	tests := []struct {
		name    string
		args    args
		want    model.Withdraw
		wantErr bool
	}{
		{"default", args{}, model.Withdraw{}, true},
		{"ref", args{ref.ID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWithdraw(db, tt.args.ID)

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("GetWithdraw() wrong Timestamp %v", got.Timestamp)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdraw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithdraw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createWithdraw(from model.AccountID, to model.AccountID, amount model.Float, batch model.BatchMode, data model.WithdrawData) model.Withdraw {
	return model.Withdraw{
		From:   from,
		To:     to,
		Amount: &amount,
		Batch:  batch,
		Data:   data,
	}
}
