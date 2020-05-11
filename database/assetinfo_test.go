// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func TestAddOrUpdateAssetInfo(t *testing.T) {
	const databaseName = "TestAddOrUpdateAssetInfo"
	t.Parallel()

	db := setup(databaseName, AssetModel())
	defer teardown(db, databaseName)

	assetInvalid := model.AssetInfo{
		AssetID:   0,
		Domain:    "foo.bar",
		Name:      "Foo Bar",
		Ticker:    "FBAR",
		Precision: 8,
	}

	assetRef := model.AssetInfo{
		AssetID:   42,
		Domain:    "foo.bar",
		Name:      "Foo Bar",
		Ticker:    "FBAR",
		Precision: 8,
	}

	type args struct {
		entry model.AssetInfo
	}
	tests := []struct {
		name    string
		args    args
		want    model.AssetInfo
		wantErr bool
	}{
		{"default", args{}, model.AssetInfo{}, true},
		{"invalid", args{assetInvalid}, model.AssetInfo{}, true},

		{"valid", args{assetRef}, assetRef, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {

			if !tt.wantErr && !tt.want.Valid() {
				t.Error("Invalid want")
			}
			got, err := AddOrUpdateAssetInfo(db, tt.args.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrUpdateAssetInfo() error = %v, wantErr %v", err, tt.wantErr)
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
				t.Errorf("AddOrUpdateAssetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAssetInfo(t *testing.T) {
	const databaseName = "TestGetAssetInfo"
	t.Parallel()

	db := setup(databaseName, AssetModel())
	defer teardown(db, databaseName)

	assetRef, _ := AddOrUpdateAssetInfo(db, model.AssetInfo{
		AssetID:   42,
		Domain:    "foo.bar",
		Name:      "Foo Bar",
		Ticker:    "FBAR",
		Precision: 8,
	})

	type args struct {
		assetID model.AssetID
	}
	tests := []struct {
		name    string
		args    args
		want    model.AssetInfo
		wantErr bool
	}{
		{"default", args{}, model.AssetInfo{}, true},
		{"notExists", args{1337}, model.AssetInfo{}, true},

		{"valid", args{assetRef.AssetID}, assetRef, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAssetInfo(db, tt.args.assetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAssetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAssetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
