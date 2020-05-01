// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func TestAddOperationInfo(t *testing.T) {
	const databaseName = "TestAddOperationInfo"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	const data = "{}"

	type args struct {
		operation model.OperationInfo
	}
	tests := []struct {
		name    string
		args    args
		want    model.OperationInfo
		wantErr bool
	}{
		{"default", args{}, model.OperationInfo{}, true},
		{"invalidUpdate", args{model.OperationInfo{ID: 1, CryptoAddressID: 42, TxID: ":txid"}}, model.OperationInfo{}, true},

		{"valid", args{model.OperationInfo{CryptoAddressID: 42, TxID: ":txid1"}}, model.OperationInfo{CryptoAddressID: 42, TxID: ":txid1"}, false},
		{"validWithAmount", args{model.OperationInfo{CryptoAddressID: 42, TxID: ":txid2", Amount: 0.1337, Data: data}}, model.OperationInfo{CryptoAddressID: 42, TxID: ":txid2", Amount: 0.1337, Data: data}, false},
		{"validWithData", args{model.OperationInfo{CryptoAddressID: 42, TxID: ":txid3", Data: data}}, model.OperationInfo{CryptoAddressID: 42, TxID: ":txid3", Data: data}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddOperationInfo(db, tt.args.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOperationInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// copy new db fields
			tt.want.ID = got.ID
			tt.want.Timestamp = got.Timestamp
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddOperationInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGetOperationInfo(t *testing.T) {
	const databaseName = "TestGetOperationInfo"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	const cryptoAddressID = model.CryptoAddressID(42)
	ref1, _ := AddOperationInfo(db, model.OperationInfo{CryptoAddressID: cryptoAddressID, TxID: ":txid1"})
	ref2, _ := AddOperationInfo(db, model.OperationInfo{CryptoAddressID: cryptoAddressID, TxID: ":txid2", Amount: 0.1337})

	type args struct {
		operationID model.OperationInfoID
	}
	tests := []struct {
		name    string
		args    args
		want    model.OperationInfo
		wantErr bool
	}{
		{"default", args{}, model.OperationInfo{}, true},
		{"notExists", args{1337}, model.OperationInfo{}, true},

		{"valid", args{ref1.ID}, ref1, false},
		{"validWithAmount", args{ref2.ID}, ref2, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOperationInfo(db, tt.args.operationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOperationInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOperationInfoByTxId(t *testing.T) {
	const databaseName = "TestGetOperationInfoByTxId"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	const cryptoAddressID = model.CryptoAddressID(42)
	ref1, _ := AddOperationInfo(db, model.OperationInfo{CryptoAddressID: cryptoAddressID, TxID: ":txid1"})

	type args struct {
		txID model.TxID
	}
	tests := []struct {
		name    string
		args    args
		want    model.OperationInfo
		wantErr bool
	}{
		{"default", args{}, model.OperationInfo{}, true},
		{"invalidTxTd", args{":wrnongTx"}, model.OperationInfo{}, true},

		{"valid", args{ref1.TxID}, ref1, false},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOperationInfoByTxId(db, tt.args.txID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationInfoByTxId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOperationInfoByTxId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOperationInfoByCryptoAddress(t *testing.T) {
	const databaseName = "TestGetOperationInfoByCryptoAddress"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	const cryptoAddressID = model.CryptoAddressID(42)
	ref1, _ := AddOperationInfo(db, model.OperationInfo{CryptoAddressID: cryptoAddressID, TxID: ":txid1"})
	ref2, _ := AddOperationInfo(db, model.OperationInfo{CryptoAddressID: cryptoAddressID, TxID: ":txid2"})
	ref3, _ := AddOperationInfo(db, model.OperationInfo{CryptoAddressID: cryptoAddressID, TxID: ":txid3"})

	type args struct {
		cryptoAddressID model.CryptoAddressID
	}
	tests := []struct {
		name    string
		args    args
		want    []model.OperationInfo
		wantErr bool
	}{
		{"default", args{}, nil, true},
		{"notExists", args{1337}, nil, false},

		{"valid", args{cryptoAddressID}, createOperationInfoList(ref1, ref2, ref3), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOperationInfoByCryptoAddress(db, tt.args.cryptoAddressID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperationInfoByCryptoAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOperationInfoByCryptoAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createOperationInfoList(operations ...model.OperationInfo) []model.OperationInfo {
	var result []model.OperationInfo
	return append(result, operations...)
}

func TestFindCryptoAddressesByOperationInfoState(t *testing.T) {
	const databaseName = "TestFindCryptoAddressesByOperationInfoState"
	t.Parallel()

	db := setup(databaseName, OperationInfoModel())
	defer teardown(db, databaseName)

	type args struct {
		chain model.String
		state model.String
	}
	tests := []struct {
		name    string
		args    args
		want    []model.CryptoAddress
		wantErr bool
	}{
		{"default", args{}, nil, true},
		{"validEmpty", args{"chain", "state"}, nil, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindCryptoAddressesByOperationInfoState(db, tt.args.chain, tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindCryptoAddressesByOperationInfoState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindCryptoAddressesByOperationInfoState() = %v, want %v", got, tt.want)
			}
		})
	}
}
