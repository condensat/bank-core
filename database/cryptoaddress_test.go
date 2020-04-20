// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/database/model"
)

func TestAddOrUpdateCryptoAddress(t *testing.T) {
	const databaseName = "TestAddOrUpdateCryptoAddress"
	t.Parallel()

	db := setup(databaseName, CryptoAddressModel())
	defer teardown(db, databaseName)

	// create db entry for duplicate test
	existingPublicAddress := model.String("bar bar")
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 100, PublicAddress: existingPublicAddress})

	type args struct {
		address model.CryptoAddress
	}
	tests := []struct {
		name    string
		args    args
		want    model.CryptoAddress
		wantErr bool
	}{
		{"default", args{}, model.CryptoAddress{}, true},

		{"invalidAccountID", args{model.CryptoAddress{PublicAddress: "foo"}}, model.CryptoAddress{}, true},
		{"invalidPublicAddress", args{model.CryptoAddress{AccountID: 42}}, model.CryptoAddress{}, true},
		{"withPublicAddress", args{model.CryptoAddress{AccountID: 1337, PublicAddress: "foo"}}, model.CryptoAddress{AccountID: 1337, PublicAddress: "foo"}, false},

		{"DuplicatesPublicAddress", args{model.CryptoAddress{AccountID: 101, PublicAddress: existingPublicAddress}}, model.CryptoAddress{}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddOrUpdateCryptoAddress(db, tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrUpdateCryptoAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// skip update tests of no entry was created
			if got.AccountID == 0 {
				return
			}

			if got.CreationDate == nil {
				t.Errorf("CreationDate should not be nil")
				return
			}

			{
				want := cloneCryptoAddress(tt.want)
				creationDate := *got.CreationDate
				want.CreationDate = &creationDate // set CreationDate for DeepEqual
				if !reflect.DeepEqual(got, want) {
					t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", got, want)
				}
			}

			ref, _ := GetCryptoAddressByAccountID(db, got.AccountID)
			checkCryptoAddressUpdate(t, db, ref)
		})
	}
}

func TestGetCryptoAddressByAccountID(t *testing.T) {
	const databaseName = "TestGetCryptoAddressByAccountID"
	t.Parallel()

	db := setup(databaseName, CryptoAddressModel())
	defer teardown(db, databaseName)

	refCryptoAddress, _ := AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 42, PublicAddress: "ref"})

	type args struct {
		accountID model.AccountID
	}
	tests := []struct {
		name    string
		args    args
		want    model.CryptoAddress
		wantErr bool
	}{
		{"Default", args{}, model.CryptoAddress{}, true},
		{"NotFound", args{1337}, model.CryptoAddress{}, true},

		{"Valid", args{refCryptoAddress.AccountID}, refCryptoAddress, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCryptoAddressByAccountID(db, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCryptoAddressByAccountID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCryptoAddressByAccountID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func checkCryptoAddressUpdate(t *testing.T, db bank.Database, ref model.CryptoAddress) {
	// fetch from db
	{
		got, err := GetCryptoAddressByAccountID(db, ref.AccountID)
		if err != nil {
			t.Errorf("GetCryptoAddressByAccountID() error= %v", err)
		}
		if !reflect.DeepEqual(got, ref) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", got, ref)
		}
	}

	// do not change CreationDate
	{
		want := cloneCryptoAddress(ref)
		cpy := cloneCryptoAddress(ref)

		timestamp := time.Now().UTC().Truncate(time.Second).Add(3 * time.Second)
		cpy.CreationDate = &timestamp

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err != nil {
			t.Errorf("AddOrUpdateCryptoAddress() error= %v", err)
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}

		check := cloneCryptoAddress(ref)
		if !reflect.DeepEqual(check, update) {
			t.Errorf("CreationDate change  stored = %+v, want %+v", check, update)
		}
	}

	// change PublicAddress
	{
		want := cloneCryptoAddress(ref)
		cpy := cloneCryptoAddress(ref)

		cpy.PublicAddress = model.String(randSeq(4))

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err != nil {
			t.Errorf("AddOrUpdateCryptoAddress() error= %v", err)
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}

		check := cloneCryptoAddress(ref)
		if !reflect.DeepEqual(check, update) {
			t.Errorf("CreationDate change  stored = %+v, want %+v", check, update)
		}
	}

	// do not revert PublicAddress to empty
	{
		want := model.CryptoAddress{}
		cpy := cloneCryptoAddress(ref)

		cpy.PublicAddress = ""

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err == nil {
			t.Errorf("AddOrUpdateCryptoAddress() should fails")
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}
	}

	// reset to reference state
	_, err := AddOrUpdateCryptoAddress(db, ref)
	if err != nil {
		t.Errorf("Failed to reset to referecnce state() error= %v", err)
	}
}

func cloneCryptoAddress(address model.CryptoAddress) model.CryptoAddress {
	result := model.CryptoAddress{
		AccountID:     address.AccountID,
		PublicAddress: address.PublicAddress,
		FirstBlockId:  address.FirstBlockId,
	}

	if address.CreationDate != nil {
		creationDate := *address.CreationDate
		result.CreationDate = &creationDate
	}

	return result
}
