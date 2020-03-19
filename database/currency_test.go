// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func Test_currencyColumnNames(t *testing.T) {
	fields := getSortedTypeFileds(reflect.TypeOf(model.Currency{}))
	names := currencyColumnNames()
	sort.Strings(names)

	if !reflect.DeepEqual(names, fields) {
		t.Errorf("columnsNames() = %v, want %v", names, fields)
	}
}

func TestCurrency(t *testing.T) {
	const databaseName = "TestAddCurrency"
	t.Parallel()

	ctx := setup(context.Background(), databaseName, CurrencyModel())
	defer teardown(ctx, databaseName)

	entries := createTestData()

	// check if table is empty
	if count := CountCurrencies(ctx); count != 0 {
		t.Errorf("Missing CountCurrencies() = %+v, want %+v", count, 0)
	}
	defer checkFinalState(t, ctx, entries)

	type args struct {
		ctx      context.Context
		currency model.Currency
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NotAvailable", args{ctx, entries[0]}, false},
		{"Available", args{ctx, entries[1]}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {

			{
				// Create Tests
				got, err := AddOrUpdateCurrency(tt.args.ctx, tt.args.currency)
				if (err != nil) != tt.wantErr {
					t.Errorf("AddOrUpdateCurrency() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(got, tt.args.currency) {
					t.Errorf("GetCurrency() = %+v, want %+v", got, tt.args.currency)
				}

				got, err = GetCurrencyByName(ctx, tt.args.currency.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				if !reflect.DeepEqual(got, tt.args.currency) {
					t.Errorf("GetCurrencyByName() = %+v, want %+v", got, tt.args.currency)
				}
			}

			// Update Tests
			{
				updateCurr, err := GetCurrencyByName(ctx, tt.args.currency.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				// change entry
				*updateCurr.Available = 2

				got, err := AddOrUpdateCurrency(tt.args.ctx, updateCurr)
				if (err != nil) != tt.wantErr {
					t.Errorf("AddOrUpdateCurrency() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(got, updateCurr) {
					t.Errorf("AddOrUpdateCurrency() = %+v, want %+v", got, updateCurr)
				}

				got, err = GetCurrencyByName(ctx, updateCurr.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				if !reflect.DeepEqual(got, updateCurr) {
					t.Errorf("GetCurrencyByName() = %+v, want %+v", got, updateCurr)
				}

				updateCurr, err = GetCurrencyByName(ctx, tt.args.currency.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				// restore entry
				*updateCurr.Available = *tt.args.currency.Available

				_, err = AddOrUpdateCurrency(tt.args.ctx, updateCurr)
				if err != nil {
					t.Errorf("WTF")
				}
				got, err = GetCurrencyByName(ctx, updateCurr.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				if !reflect.DeepEqual(got, updateCurr) {
					t.Errorf("GetCurrencyByName() = %+v, want %+v", got, updateCurr)
				}
			}

		})
	}
}

func createTestData() []model.Currency {
	return []model.Currency{
		model.NewCurrency("USD", 0),
		model.NewCurrency("BTC", 1),
	}
}

func checkFinalState(t *testing.T, ctx context.Context, entries []model.Currency) {
	// check if table has entries
	if count := CountCurrencies(ctx); count != len(entries) {
		t.Errorf("Missing CountCurrencies() = %+v, want %+v", count, len(entries))
	}

	{
		list, err := ListAllCurrency(ctx)
		if err != nil {
			t.Errorf("ListAllCurrency() Failed = %+v", err)
		}
		if len(list) != len(entries) {
			t.Errorf("Missing ListAllCurrency() = %+v, want %+v", len(list), len(entries))
		}
	}

	{
		list, err := ListAvailableCurrency(ctx)
		if err != nil {
			t.Errorf("ListAvailableCurrency() Failed = %+v", err)
		}
		if len(list) != len(entries)/2 {
			t.Errorf("Missing ListAvailableCurrency() = %+v, want %+v", len(list), len(entries)/2)
		}

		for _, curr := range list {
			if !curr.IsAvailable() {
				t.Errorf("Currency IsAvailable must be true")
			}
		}
	}
}
