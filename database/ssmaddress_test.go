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

func TestAddSsmAddress(t *testing.T) {
	const databaseName = "TestAddSsmAddress"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	type args struct {
		address model.SsmAddress
		info    model.SsmAddressInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"default", args{}, true},

		{"valid", args{address, info}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddSsmAddress(db, tt.args.address, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSsmAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == 0 != tt.wantErr {
				t.Errorf("AddSsmAddress() = invalid id %v", got)
			}
		})
	}
}

func TestCountSsmAddress(t *testing.T) {
	const databaseName = "TestCountSsmAddress"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	_, _ = AddSsmAddress(db, address, info)
	type args struct {
		chain       model.SsmChain
		fingerprint model.SsmFingerprint
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"default", args{}, 0, true},
		{"invalidchain", args{"", info.Fingerprint}, 0, true},
		{"invalidfingerprint", args{info.Chain, ""}, 0, true},

		{"zero", args{info.Chain, "other"}, 0, false},
		{"one", args{info.Chain, info.Fingerprint}, 1, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountSsmAddress(db, tt.args.chain, tt.args.fingerprint)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountSsmAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountSsmAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountSsmAddressByState(t *testing.T) {
	const databaseName = "TestCountSsmAddressByState"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	refID, err := AddSsmAddress(db, address, info)
	if err != nil {
		t.Errorf("AddSsmAddress failed. %s", err)
		return
	}

	_, err = UpdateSsmAddressState(db, refID, model.SsmAddressStatusUsed)
	if err != nil {
		t.Errorf("UpdateSsmAddressState failed. %s", err)
		return
	}

	type args struct {
		chain       model.SsmChain
		fingerprint model.SsmFingerprint
		state       model.SsmAddressStatus
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"default", args{}, 0, true},
		{"unused", args{info.Chain, info.Fingerprint, model.SsmAddressStatusUsed}, 1, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountSsmAddressByState(db, tt.args.chain, tt.args.fingerprint, tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountSsmAddressByState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountSsmAddressByState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSsmAddress(t *testing.T) {
	const databaseName = "TestGetSsmAddress"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	refID, _ := AddSsmAddress(db, address, info)
	ref := model.SsmAddress{
		ID:            refID,
		PublicAddress: address.PublicAddress,
		ScriptPubkey:  address.ScriptPubkey,
		BlindingKey:   address.BlindingKey,
	}

	type args struct {
		addressID model.SsmAddressID
	}
	tests := []struct {
		name    string
		args    args
		want    model.SsmAddress
		wantErr bool
	}{
		{"default", args{}, model.SsmAddress{}, true},

		{"ref", args{refID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSsmAddress(db, tt.args.addressID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSsmAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSsmAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSsmAddressInfo(t *testing.T) {
	const databaseName = "TestGetSsmAddressInfo"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	refID, _ := AddSsmAddress(db, address, info)
	ref := model.SsmAddressInfo{
		SsmAddressID: refID,
		Chain:        info.Chain,
		Fingerprint:  info.Fingerprint,
		HDPath:       info.HDPath,
	}

	type args struct {
		addressID model.SsmAddressID
	}
	tests := []struct {
		name    string
		args    args
		want    model.SsmAddressInfo
		wantErr bool
	}{
		{"default", args{}, model.SsmAddressInfo{}, true},

		{"ref", args{refID}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSsmAddressInfo(db, tt.args.addressID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSsmAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSsmAddressInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSsmAddressByPublicAddress(t *testing.T) {
	const databaseName = "TestGetSsmAddressByPublicAddress"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	refID, _ := AddSsmAddress(db, address, info)
	ref := model.SsmAddress{
		ID:            refID,
		PublicAddress: address.PublicAddress,
		ScriptPubkey:  address.ScriptPubkey,
		BlindingKey:   address.BlindingKey,
	}

	type args struct {
		publicAddress model.SsmPublicAddress
	}
	tests := []struct {
		name    string
		args    args
		want    model.SsmAddress
		wantErr bool
	}{
		{"default", args{}, model.SsmAddress{}, true},

		{"ref", args{address.PublicAddress}, ref, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSsmAddressByPublicAddress(db, tt.args.publicAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSsmAddressByPublicAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSsmAddressByPublicAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSsmAddressState(t *testing.T) {
	const databaseName = "TestGetSsmAddressState"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	refID, _ := AddSsmAddress(db, address, info)
	refUnused := model.SsmAddressState{
		ID:           0,
		SsmAddressID: refID,
		State:        model.SsmAddressStatusUnused,
	}

	type args struct {
		addressID model.SsmAddressID
	}
	tests := []struct {
		name    string
		args    args
		want    model.SsmAddressState
		wantErr bool
	}{
		{"default", args{}, model.SsmAddressState{}, true},

		{"ref", args{refID}, refUnused, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSsmAddressState(db, tt.args.addressID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSsmAddressState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got.ID = 0
			got.Timestamp = time.Time{}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSsmAddressState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateSsmAddressState(t *testing.T) {
	const databaseName = "TestUpdateSsmAddressState"
	t.Parallel()

	db := setup(databaseName, SsmAddressModel())
	defer teardown(db, databaseName)

	address := model.SsmAddress{ID: 0, PublicAddress: "foo", ScriptPubkey: "bar", BlindingKey: "foobar"}
	info := model.SsmAddressInfo{SsmAddressID: 42, Chain: "chain", Fingerprint: "ffffffff", HDPath: "path"}

	refID, _ := AddSsmAddress(db, address, info)
	refUsed := model.SsmAddressState{
		ID:           0,
		SsmAddressID: refID,
		State:        model.SsmAddressStatusUsed,
	}

	type args struct {
		addressID model.SsmAddressID
		status    model.SsmAddressStatus
	}
	tests := []struct {
		name    string
		args    args
		want    model.SsmAddressState
		wantErr bool
	}{
		{"default", args{}, model.SsmAddressState{}, true},

		{"ref", args{refID, model.SsmAddressStatusUsed}, refUsed, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateSsmAddressState(db, tt.args.addressID, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSsmAddressState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got.ID = 0
			got.Timestamp = time.Time{}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateSsmAddressState() = %v, want %v", got, tt.want)
			}
		})
	}
}
