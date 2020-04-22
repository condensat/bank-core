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

	const chain = model.String("chain1")

	// create db entry for duplicate test
	existingPublicAddress := model.String("bar bar")
	_, err := AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 100, PublicAddress: existingPublicAddress, Chain: chain})
	if err != nil {
		t.Errorf("failed to AddOrUpdateCryptoAddress for duplicate tests")
	}

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

		{"InvalidDuplicates", args{model.CryptoAddress{AccountID: 101, PublicAddress: existingPublicAddress, Chain: chain}}, model.CryptoAddress{}, true},
		{"invalidChain", args{model.CryptoAddress{AccountID: 1337, PublicAddress: "foo"}}, model.CryptoAddress{}, true},
		{"invalidAllChain", args{model.CryptoAddress{AccountID: 1337, PublicAddress: "foo", Chain: AllChains}}, model.CryptoAddress{}, true},

		{"valid", args{model.CryptoAddress{AccountID: 1337, PublicAddress: "foo", Chain: chain}}, model.CryptoAddress{AccountID: 1337, PublicAddress: "foo", Chain: chain}, false},
		{"validMultiple", args{model.CryptoAddress{AccountID: 1337, PublicAddress: "bar", Chain: chain}}, model.CryptoAddress{AccountID: 1337, PublicAddress: "bar", Chain: chain}, false},
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

			if got.CreationDate == nil || got.CreationDate.IsZero() {
				t.Errorf("Invalid CreationDate: %v", got.CreationDate)
				return
			}

			{
				want := cloneCryptoAddress(tt.want)
				want.ID = got.ID
				want.CreationDate = got.CreationDate // set CreationDate for DeepEqual

				if !reflect.DeepEqual(got, want) {
					t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", got, want)
				}
			}

			ref, _ := GetCryptoAddress(db, got.ID)
			checkCryptoAddressUpdate(t, db, ref)
		})
	}
}

func TestGetCryptoAddress(t *testing.T) {
	const databaseName = "TestGetCryptoAddress"
	t.Parallel()

	db := setup(databaseName, CryptoAddressModel())
	defer teardown(db, databaseName)

	const chain = model.String("chain1")

	accountID := model.AccountID(42)
	ref1, _ := AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref1", Chain: chain})

	type args struct {
		ID model.ID
	}
	tests := []struct {
		name    string
		args    args
		want    model.CryptoAddress
		wantErr bool
	}{
		{"empty", args{}, model.CryptoAddress{}, true},
		{"notFound", args{42}, model.CryptoAddress{}, true},
		{"ref1", args{ref1.ID}, cloneCryptoAddress(ref1), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCryptoAddress(db, tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCryptoAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCryptoAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastAccountCryptoAddress(t *testing.T) {
	const databaseName = "TestLastAccountCryptoAddress"
	t.Parallel()

	db := setup(databaseName, CryptoAddressModel())
	defer teardown(db, databaseName)

	const chain = model.String("chain1")

	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 42, PublicAddress: "ref1", Chain: chain})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 42, PublicAddress: "ref2", Chain: chain})
	lastRef, _ := AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 42, PublicAddress: "ref3", Chain: chain})

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
		{"NotFound", args{1337}, model.CryptoAddress{}, false},

		{"Valid", args{lastRef.AccountID}, lastRef, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := LastAccountCryptoAddress(db, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LastAccountCryptoAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want := cloneCryptoAddress(tt.want)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("LastAccountCryptoAddress() = %v, want %v", got, want)
			}
		})
	}
}

func TestAllAccountCryptoAddresses(t *testing.T) {
	const databaseName = "TestAllAccountCryptoAddresses"
	t.Parallel()

	db := setup(databaseName, CryptoAddressModel())
	defer teardown(db, databaseName)

	const chain = model.String("chain1")

	accountID := model.AccountID(42)
	ref1, _ := AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref1", Chain: chain})
	ref2, _ := AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref2", Chain: chain})
	ref3, _ := AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref3", Chain: chain})
	allRefs := []model.CryptoAddress{
		ref1,
		ref2,
		ref3,
	}

	type args struct {
		accountID model.AccountID
	}
	tests := []struct {
		name    string
		args    args
		want    []model.CryptoAddress
		wantErr bool
	}{
		{"Default", args{}, nil, true},
		{"NotFound", args{1337}, nil, false},

		{"Valid", args{accountID}, allRefs, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AllAccountCryptoAddresses(db, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LastAccountCryptoAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastAccountCryptoAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllUnusedAccountCryptoAddresses(t *testing.T) {
	const databaseName = "TestAllUnusedAccountCryptoAddresses"
	t.Parallel()

	db := setup(databaseName, CryptoAddressModel())
	defer teardown(db, databaseName)

	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 42, PublicAddress: "ref1", Chain: "chain1", FirstBlockId: 424242})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 42, PublicAddress: "ref2", Chain: "chain1", FirstBlockId: 0})

	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 1337, PublicAddress: "ref3", Chain: "chain2", FirstBlockId: 1})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 1337, PublicAddress: "ref4", Chain: "chain2", FirstBlockId: 0})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: 1337, PublicAddress: "ref5", Chain: "chain2", FirstBlockId: 0})

	type args struct {
		accountID model.AccountID
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"default", args{0}, 0, true},

		{"account42", args{42}, 1, false},
		{"account1337", args{1337}, 2, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AllUnusedAccountCryptoAddresses(db, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllUnusedAccountCryptoAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("AllUnusedAccountCryptoAddresses() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestAllUnusedCryptoAddresses(t *testing.T) {
	const databaseName = "TestAllUnusedCryptoAddresses"
	t.Parallel()

	db := setup(databaseName, CryptoAddressModel())
	defer teardown(db, databaseName)

	accountID := model.AccountID(42)
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref1", Chain: "chain1", FirstBlockId: 424242})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref2", Chain: "chain1", FirstBlockId: 0})

	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref3", Chain: "chain2", FirstBlockId: 1})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref4", Chain: "chain2", FirstBlockId: 0})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref5", Chain: "chain2", FirstBlockId: 0})

	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref7", Chain: "chain3", FirstBlockId: 0})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref8", Chain: "chain3", FirstBlockId: 0})
	_, _ = AddOrUpdateCryptoAddress(db, model.CryptoAddress{AccountID: accountID, PublicAddress: "ref9", Chain: "chain3", FirstBlockId: 0})

	type args struct {
		chain model.String
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"default", args{""}, 0, true},
		{"allchains", args{"*"}, 6, false},

		{"chain1", args{"chain1"}, 1, false},
		{"chain2", args{"chain2"}, 2, false},
		{"chain3", args{"chain3"}, 3, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AllUnusedCryptoAddresses(db, tt.args.chain)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllUnusedCryptoAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("AllUnusedCryptoAddresses() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func checkCryptoAddressUpdate(t *testing.T, db bank.Database, ref model.CryptoAddress) {
	// fetch from db
	{
		list, err := AllAccountCryptoAddresses(db, ref.AccountID)
		if err != nil {
			t.Errorf("GetCryptoAddressByAccountID() error= %v", err)
		}
		if ok, got := containsCryptoAddress(list, ref); ok {
			if !reflect.DeepEqual(got, ref) {
				t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", got, ref)
			}
		}
	}

	// do not change CreationDate
	{
		want, _ := GetCryptoAddress(db, ref.ID)
		cpy, _ := GetCryptoAddress(db, ref.ID)

		timestamp := time.Now().UTC().Truncate(time.Second).Add(3 * time.Second)
		cpy.CreationDate = &timestamp

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err != nil {
			t.Errorf("AddOrUpdateCryptoAddress() error= %v", err)
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}

		check, _ := GetCryptoAddress(db, ref.ID)
		if !reflect.DeepEqual(check, update) {
			t.Errorf("CreationDate change stored = %+v, want %+v", check, update)
		}
	}

	// change PublicAddress
	{
		want, _ := GetCryptoAddress(db, ref.ID)
		cpy, _ := GetCryptoAddress(db, ref.ID)

		cpy.PublicAddress = model.String(randSeq(4))

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err != nil {
			t.Errorf("AddOrUpdateCryptoAddress() error= %v", err)
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}

		check, _ := GetCryptoAddress(db, ref.ID)
		if !reflect.DeepEqual(check, update) {
			t.Errorf("PublicAddress change not stored = %+v, want %+v", check, update)
		}
	}

	// do not revert PublicAddress to empty
	{
		want := model.CryptoAddress{}
		cpy, _ := GetCryptoAddress(db, ref.ID)

		cpy.PublicAddress = ""

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err == nil {
			t.Errorf("AddOrUpdateCryptoAddress() should fails")
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}
	}

	// Mempool
	{
		want, _ := GetCryptoAddress(db, ref.ID)
		cpy, _ := GetCryptoAddress(db, ref.ID)

		want.FirstBlockId = 1
		cpy.FirstBlockId = want.FirstBlockId

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err != nil {
			t.Errorf("AddOrUpdateCryptoAddress() error= %v", err)
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}

		if !update.IsUsed() {
			t.Errorf("Updated CryptoAddress should be in use: %+v, want %+v", update, want)
		}

		store, _ := GetCryptoAddress(db, ref.ID)
		if !reflect.DeepEqual(store, update) {
			t.Errorf("Mempool change not stored = %+v, want %+v", store, update)
		}
	}

	// Mined
	{
		want, _ := GetCryptoAddress(db, ref.ID)
		cpy, _ := GetCryptoAddress(db, ref.ID)

		want.FirstBlockId = 424242
		cpy.FirstBlockId = want.FirstBlockId

		update, err := AddOrUpdateCryptoAddress(db, cpy)
		if err != nil {
			t.Errorf("AddOrUpdateCryptoAddress() error= %v", err)
		}
		if !reflect.DeepEqual(update, want) {
			t.Errorf("AddOrUpdateCryptoAddress() = %+v, want %+v", update, want)
		}

		if !update.IsUsed() {
			t.Errorf("Updated CryptoAddress should be in use: %+v, want %+v", update, want)
		}

		if update.Confirmations(424242) != 1 {
			t.Errorf("Failed to update FirstBlockId: %+v, want %+v", update, want)
		}

		store, _ := GetCryptoAddress(db, update.ID)
		if !reflect.DeepEqual(store, update) {
			t.Errorf("Mined change not stored = %+v, want %+v", store, update)
		}
	}

	// reset to reference state
	_, err := AddOrUpdateCryptoAddress(db, ref)
	if err != nil {
		t.Errorf("Failed to reset to referecnce state() error= %v", err)
	}
}

func cloneCryptoAddress(address model.CryptoAddress) model.CryptoAddress {
	result := address

	if address.CreationDate != nil {
		creationDate := *address.CreationDate
		result.CreationDate = &creationDate
	}
	return result
}

func containsCryptoAddress(list []model.CryptoAddress, item model.CryptoAddress) (bool, model.CryptoAddress) {
	for _, address := range list {
		if address.ID == item.ID {
			return true, address
		}
	}
	return false, model.CryptoAddress{}
}
