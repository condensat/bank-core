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

func TestAddOrUpdateAssetIcon(t *testing.T) {
	const databaseName = "TestAddOrUpdateAssetIcon"
	t.Parallel()

	db := tests.Setup(databaseName, AssetModel())
	defer tests.Teardown(db, databaseName)

	assetNil := model.AssetIcon{
		AssetID: 40,
		Data:    nil,
	}
	assetZero := model.AssetIcon{
		AssetID: 41,
		Data:    []byte{},
	}
	assetRef := model.AssetIcon{
		AssetID: 42,
		Data:    []byte("I'm alive"),
	}
	assetBig := model.AssetIcon{
		AssetID: 1337,
		Data:    make([]byte, MaxAssetIconDataLen),
	}
	assetTooBig := model.AssetIcon{
		AssetID: 1338,
		Data:    make([]byte, MaxAssetIconDataLen+1),
	}

	type args struct {
		entry model.AssetIcon
	}
	tests := []struct {
		name    string
		args    args
		want    model.AssetIcon
		wantErr bool
	}{
		{"default", args{}, model.AssetIcon{}, true},

		{"invalidNil", args{assetNil}, model.AssetIcon{}, true},
		{"invalidZero", args{assetZero}, model.AssetIcon{}, true},
		{"invalidTooBig", args{assetTooBig}, model.AssetIcon{}, true},

		{"valid", args{assetRef}, assetRef, false},
		{"validBig", args{assetBig}, assetBig, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddOrUpdateAssetIcon(db, tt.args.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrUpdateAssetIcon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.LastUpdate.IsZero() {
					t.Errorf("Invalid LastUpdate timestamp. %v", assetRef.LastUpdate)
				}
				// reset timestamp for DeepEqual
				got.LastUpdate = assetRef.LastUpdate
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddOrUpdateAssetIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAssetIcon(t *testing.T) {
	const databaseName = "TestGetAssetIcon"
	t.Parallel()

	db := tests.Setup(databaseName, AssetModel())
	defer tests.Teardown(db, databaseName)

	assetRef, _ := AddOrUpdateAssetIcon(db, model.AssetIcon{
		AssetID: 42,
		Data:    []byte("I'm alive"),
	})

	type args struct {
		assetID model.AssetID
	}
	tests := []struct {
		name    string
		args    args
		want    model.AssetIcon
		wantErr bool
	}{
		{"default", args{}, model.AssetIcon{}, true},
		{"notExists", args{1337}, model.AssetIcon{}, true},

		{"valid", args{assetRef.AssetID}, assetRef, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAssetIcon(db, tt.args.assetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAssetIcon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAssetIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}
