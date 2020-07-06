// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
)

func TestFromSepaData(t *testing.T) {
	t.Parallel()

	ref := WithdrawTargetSepaData{
		WithdrawTargetFiatData: WithdrawTargetFiatData{Network: "sepa"},

		BIC:  "BIC",
		IBAN: "IBAN",
	}
	type args struct {
		withdrawID WithdrawID
		data       WithdrawTargetSepaData
	}
	tests := []struct {
		name string
		args args
		want WithdrawTargetSepaData
	}{
		{"default", args{}, WithdrawTargetSepaData{}},
		{"valid", args{42, ref}, ref},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := FromSepaData(tt.args.withdrawID, tt.args.data)

			data, _ := got.SepaData()
			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("FromSepaData() = %v, want %v", data, tt.want)
			}
		})
	}
}

func TestWithdrawTarget_SepaData(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID         WithdrawTargetID
		WithdrawID WithdrawID
		Type       WithdrawTargetType
		Data       WithdrawTargetData
	}
	tests := []struct {
		name    string
		fields  fields
		want    WithdrawTargetSepaData
		wantErr bool
	}{
		{"default", fields{}, WithdrawTargetSepaData{}, true},
		{"type", fields{Type: WithdrawTargetSepa}, WithdrawTargetSepaData{}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &WithdrawTarget{
				ID:         tt.fields.ID,
				WithdrawID: tt.fields.WithdrawID,
				Type:       tt.fields.Type,
				Data:       tt.fields.Data,
			}
			got, err := p.SepaData()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawTarget.SepaData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithdrawTarget.SepaData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromSwiftData(t *testing.T) {
	t.Parallel()

	ref := WithdrawTargetSwiftData{
		WithdrawTargetFiatData: WithdrawTargetFiatData{Network: "swift"},

		CountryCode: "CHE",
		Bank:        "Condensat",
		Account:     "1337",
	}
	type args struct {
		withdrawID WithdrawID
		data       WithdrawTargetSwiftData
	}
	tests := []struct {
		name string
		args args
		want WithdrawTargetSwiftData
	}{
		{"default", args{}, WithdrawTargetSwiftData{}},
		{"valid", args{42, ref}, ref},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := FromSwiftData(tt.args.withdrawID, tt.args.data)

			data, _ := got.SwiftData()
			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("FromSwiftData() = %v, want %v", data, tt.want)
			}
		})
	}
}

func TestWithdrawTarget_SwiftData(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID         WithdrawTargetID
		WithdrawID WithdrawID
		Type       WithdrawTargetType
		Data       WithdrawTargetData
	}
	tests := []struct {
		name    string
		fields  fields
		want    WithdrawTargetSwiftData
		wantErr bool
	}{
		{"default", fields{}, WithdrawTargetSwiftData{}, true},
		{"type", fields{Type: WithdrawTargetSwift}, WithdrawTargetSwiftData{}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &WithdrawTarget{
				ID:         tt.fields.ID,
				WithdrawID: tt.fields.WithdrawID,
				Type:       tt.fields.Type,
				Data:       tt.fields.Data,
			}
			got, err := p.SwiftData()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawTarget.SwiftData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithdrawTarget.SwiftData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromCardData(t *testing.T) {
	t.Parallel()

	ref := WithdrawTargetCardData{
		WithdrawTargetFiatData: WithdrawTargetFiatData{Network: "card"},

		PAN: "4222222222222",
	}
	type args struct {
		withdrawID WithdrawID
		data       WithdrawTargetCardData
	}
	tests := []struct {
		name string
		args args
		want WithdrawTargetCardData
	}{
		{"default", args{}, WithdrawTargetCardData{}},
		{"valid", args{42, ref}, ref},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := FromCardData(tt.args.withdrawID, tt.args.data)

			data, _ := got.CardData()
			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("FromCardData() = %v, want %v", data, tt.want)
			}
		})
	}
}

func TestWithdrawTarget_CardData(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID         WithdrawTargetID
		WithdrawID WithdrawID
		Type       WithdrawTargetType
		Data       WithdrawTargetData
	}
	tests := []struct {
		name    string
		fields  fields
		want    WithdrawTargetCardData
		wantErr bool
	}{
		{"default", fields{}, WithdrawTargetCardData{}, true},
		{"type", fields{Type: WithdrawTargetCard}, WithdrawTargetCardData{}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &WithdrawTarget{
				ID:         tt.fields.ID,
				WithdrawID: tt.fields.WithdrawID,
				Type:       tt.fields.Type,
				Data:       tt.fields.Data,
			}
			got, err := p.CardData()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawTarget.CardData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithdrawTarget.CardData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromOnChainData(t *testing.T) {
	t.Parallel()

	ref := WithdrawTargetOnChainData{
		WithdrawTargetCryptoData: WithdrawTargetCryptoData{Chain: "bitcoin", PublicKey: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa "},
	}
	type args struct {
		withdrawID WithdrawID
		data       WithdrawTargetOnChainData
	}
	tests := []struct {
		name string
		args args
		want WithdrawTargetOnChainData
	}{
		{"default", args{}, WithdrawTargetOnChainData{}},
		{"valid", args{42, ref}, ref},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := FromOnChainData(tt.args.withdrawID, "bitcoin", tt.args.data)

			data, _ := got.OnChainData()
			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("FromOnChainData() = %v, want %v", data, tt.want)
			}
		})
	}
}

func TestWithdrawTarget_OnChainData(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID         WithdrawTargetID
		WithdrawID WithdrawID
		Type       WithdrawTargetType
		Data       WithdrawTargetData
	}
	tests := []struct {
		name    string
		fields  fields
		want    WithdrawTargetOnChainData
		wantErr bool
	}{
		{"default", fields{}, WithdrawTargetOnChainData{}, true},
		{"type", fields{Type: WithdrawTargetOnChain}, WithdrawTargetOnChainData{}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &WithdrawTarget{
				ID:         tt.fields.ID,
				WithdrawID: tt.fields.WithdrawID,
				Type:       tt.fields.Type,
				Data:       tt.fields.Data,
			}
			got, err := p.OnChainData()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawTarget.OnChainData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithdrawTarget.OnChainData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromLiquidData(t *testing.T) {
	t.Parallel()

	ref := WithdrawTargetLiquidData{
		WithdrawTargetCryptoData: WithdrawTargetCryptoData{Chain: "liquid", PublicKey: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa "},
	}
	type args struct {
		withdrawID WithdrawID
		data       WithdrawTargetLiquidData
	}
	tests := []struct {
		name string
		args args
		want WithdrawTargetLiquidData
	}{
		{"default", args{}, WithdrawTargetLiquidData{}},
		{"valid", args{42, ref}, ref},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := FromLiquidData(tt.args.withdrawID, tt.args.data)

			data, _ := got.LiquidData()
			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("FromLiquidData() = %v, want %v", data, tt.want)
			}
		})
	}
}

func TestWithdrawTarget_LiquidData(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID         WithdrawTargetID
		WithdrawID WithdrawID
		Type       WithdrawTargetType
		Data       WithdrawTargetData
	}
	tests := []struct {
		name    string
		fields  fields
		want    WithdrawTargetLiquidData
		wantErr bool
	}{
		{"default", fields{}, WithdrawTargetLiquidData{}, true},
		{"type", fields{Type: WithdrawTargetLiquid}, WithdrawTargetLiquidData{}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &WithdrawTarget{
				ID:         tt.fields.ID,
				WithdrawID: tt.fields.WithdrawID,
				Type:       tt.fields.Type,
				Data:       tt.fields.Data,
			}
			got, err := p.LiquidData()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawTarget.LiquidData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithdrawTarget.LiquidData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromLightningData(t *testing.T) {
	t.Parallel()

	ref := WithdrawTargetLightningData{
		WithdrawTargetCryptoData: WithdrawTargetCryptoData{Chain: "lightning", PublicKey: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa "},
	}
	type args struct {
		withdrawID WithdrawID
		data       WithdrawTargetLightningData
	}
	tests := []struct {
		name string
		args args
		want WithdrawTargetLightningData
	}{
		{"default", args{}, WithdrawTargetLightningData{}},
		{"valid", args{42, ref}, ref},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := FromLightningData(tt.args.withdrawID, tt.args.data)

			data, _ := got.LightningData()
			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("FromLightningData() = %v, want %v", data, tt.want)
			}
		})
	}
}

func TestWithdrawTarget_LightningData(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID         WithdrawTargetID
		WithdrawID WithdrawID
		Type       WithdrawTargetType
		Data       WithdrawTargetData
	}
	tests := []struct {
		name    string
		fields  fields
		want    WithdrawTargetLightningData
		wantErr bool
	}{
		{"default", fields{}, WithdrawTargetLightningData{}, true},
		{"type", fields{Type: WithdrawTargetLightning}, WithdrawTargetLightningData{}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &WithdrawTarget{
				ID:         tt.fields.ID,
				WithdrawID: tt.fields.WithdrawID,
				Type:       tt.fields.Type,
				Data:       tt.fields.Data,
			}
			got, err := p.LightningData()
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawTarget.LightningData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithdrawTarget.LightningData() = %v, want %v", got, tt.want)
			}
		})
	}
}
