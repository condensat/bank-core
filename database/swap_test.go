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

func TestAddSwap(t *testing.T) {
	const databaseName = "TestAddSwap"
	t.Parallel()

	db := setup(databaseName, SwapModel())
	defer teardown(db, databaseName)

	type args struct {
		swapType        model.SwapType
		cryptoAddressID model.CryptoAddressID
		debitAsset      model.AssetID
		debitAmount     model.Float
		creditAsset     model.AssetID
		creditAmount    model.Float
	}
	tests := []struct {
		name    string
		args    args
		want    model.Swap
		wantErr bool
	}{
		{"default", args{}, model.Swap{}, true},
		{"valid", args{model.SwapTypeAsk, 42, 101, 1.1, 102, 1.2}, createSwap(model.SwapTypeAsk, 42, 101, 1.1, 102, 1.2), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddSwap(db, tt.args.swapType, tt.args.cryptoAddressID, tt.args.debitAsset, tt.args.debitAmount, tt.args.creditAsset, tt.args.creditAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("AddSwap() wrong Timestamp %v", got.Timestamp)
				}
				if got.ValidUntil.IsZero() || got.Timestamp.After(got.ValidUntil) {
					t.Errorf("AddSwap() wrong ValidUntil: %v %v", got.Timestamp, got.ValidUntil)
				}
			}

			if !tt.wantErr {
				tt.want.ID = got.ID
				tt.want.Timestamp = got.Timestamp
				tt.want.ValidUntil = got.ValidUntil
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddSwap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSwap(t *testing.T) {
	const databaseName = "TestGetSwap"
	t.Parallel()

	db := setup(databaseName, SwapModel())
	defer teardown(db, databaseName)

	swapRef, _ := AddSwap(db, model.SwapTypeAsk, 42, 101, 1.1, 102, 1.2)

	type args struct {
		swapID model.SwapID
	}
	tests := []struct {
		name    string
		args    args
		want    model.Swap
		wantErr bool
	}{
		{"default", args{}, model.Swap{}, true},
		{"ref", args{swapRef.ID}, swapRef, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSwap(db, tt.args.swapID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSwap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createSwap(swapType model.SwapType, cryptoAddressID model.CryptoAddressID, debitAsset model.AssetID, debitAmount model.Float, creditAsset model.AssetID, creditAmount model.Float) model.Swap {
	return model.Swap{
		Type:            swapType,
		CryptoAddressID: cryptoAddressID,
		DebitAsset:      debitAsset,
		DebitAmount:     debitAmount,
		CreditAsset:     creditAsset,
		CreditAmount:    creditAmount,
	}
}
