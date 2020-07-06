// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type WithdrawTargetID ID
type WithdrawTargetType DataType
type WithdrawTargetData Data

const (
	// Fiat
	WithdrawTargetSepa  WithdrawTargetType = "sepa"
	WithdrawTargetSwift WithdrawTargetType = "swift"
	WithdrawTargetCard  WithdrawTargetType = "card"

	// Crypto
	WithdrawTargetOnChain   WithdrawTargetType = "onchain"
	WithdrawTargetLiquid    WithdrawTargetType = "liquid"
	WithdrawTargetLightning WithdrawTargetType = "lightning"
)

type WithdrawTarget struct {
	ID         WithdrawTargetID   `gorm:"primary_key"`
	WithdrawID WithdrawID         `gorm:"index;not null"`                  // [FK] Reference to Withdraw table
	Type       WithdrawTargetType `gorm:"index;not null;size:16"`          // DataType [onchain, liquid, lightning, sepa, swift, card]
	Data       WithdrawTargetData `gorm:"type:blob;not null;default:'{}'"` // WithdrawTarget data
}

// WithdrawTargetCryptoData data type for WithdrawTargetType crypto, liquid & lightning
type WithdrawTargetCryptoData struct {
	Chain     string `json:"chain,omitempty"`
	PublicKey string `json:"publickey,omitempty"`
}

// WithdrawTargetFiatData data type for WithdrawTargetType sepa, swift & cb
type WithdrawTargetFiatData struct {
	Network string `json:"network,omitempty"`
}

// WithdrawTargetSepaData data type for WithdrawTargetType sepa
type WithdrawTargetSepaData struct {
	WithdrawTargetFiatData
	BIC  string `json:"bic,omitempty"`
	IBAN string `json:"iban,omitempty"`
}

func FromSepaData(withdrawID WithdrawID, sepa WithdrawTargetSepaData) WithdrawTarget {
	if withdrawID > 0 {
		sepa.WithdrawTargetFiatData.Network = "sepa"
	}
	data, _ := EncodeData(&sepa)
	return WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       WithdrawTargetSepa,
		Data:       WithdrawTargetData(data),
	}
}

func (p *WithdrawTarget) SepaData() (WithdrawTargetSepaData, error) {
	switch p.Type {

	case WithdrawTargetSepa:
		var data WithdrawTargetSepaData
		err := DecodeData(&data, Data(p.Data))
		return data, err

	default:
		return WithdrawTargetSepaData{}, ErrInvalidDataType
	}
}

// WithdrawTargetSwiftData data type for WithdrawTargetType sepa
type WithdrawTargetSwiftData struct {
	WithdrawTargetFiatData
	CountryCode string `json:"country_code,omitempty"`
	Bank        string `json:"bank,omitempty"`
	Account     string `json:"account,omitempty"`
}

func FromSwiftData(withdrawID WithdrawID, swift WithdrawTargetSwiftData) WithdrawTarget {
	if withdrawID > 0 {
		swift.WithdrawTargetFiatData.Network = "swift"
	}
	data, _ := EncodeData(&swift)
	return WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       WithdrawTargetSwift,
		Data:       WithdrawTargetData(data),
	}
}

func (p *WithdrawTarget) SwiftData() (WithdrawTargetSwiftData, error) {
	switch p.Type {

	case WithdrawTargetSwift:
		var data WithdrawTargetSwiftData
		err := DecodeData(&data, Data(p.Data))
		return data, err

	default:
		return WithdrawTargetSwiftData{}, ErrInvalidDataType
	}
}

// WithdrawTargetCardData data type for WithdrawTargetType card
type WithdrawTargetCardData struct {
	WithdrawTargetFiatData
	PAN string `json:"pan,omitempty"`
}

func FromCardData(withdrawID WithdrawID, card WithdrawTargetCardData) WithdrawTarget {
	if withdrawID > 0 {
		card.WithdrawTargetFiatData.Network = "card"
	}
	data, _ := EncodeData(&card)
	return WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       WithdrawTargetCard,
		Data:       WithdrawTargetData(data),
	}
}

func (p *WithdrawTarget) CardData() (WithdrawTargetCardData, error) {
	switch p.Type {

	case WithdrawTargetCard:
		var data WithdrawTargetCardData
		err := DecodeData(&data, Data(p.Data))
		return data, err

	default:
		return WithdrawTargetCardData{}, ErrInvalidDataType
	}
}

// WithdrawTargetOnChainData data type for WithdrawTargetType crypto
type WithdrawTargetOnChainData struct {
	WithdrawTargetCryptoData
}

func FromOnChainData(withdrawID WithdrawID, chain string, onChain WithdrawTargetOnChainData) WithdrawTarget {
	if withdrawID > 0 {
		onChain.WithdrawTargetCryptoData.Chain = chain
	}
	data, _ := EncodeData(&onChain)
	return WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       WithdrawTargetOnChain,
		Data:       WithdrawTargetData(data),
	}
}

func (p *WithdrawTarget) OnChainData() (WithdrawTargetOnChainData, error) {
	switch p.Type {

	case WithdrawTargetOnChain:
		var data WithdrawTargetOnChainData
		err := DecodeData(&data, Data(p.Data))
		return data, err

	default:
		return WithdrawTargetOnChainData{}, ErrInvalidDataType
	}
}

// WithdrawTargetLiquidData data type for WithdrawTargetType liquid
type WithdrawTargetLiquidData struct {
	WithdrawTargetCryptoData
}

func FromLiquidData(withdrawID WithdrawID, liquid WithdrawTargetLiquidData) WithdrawTarget {
	if withdrawID > 0 {
		liquid.WithdrawTargetCryptoData.Chain = "liquid"
	}
	data, _ := EncodeData(&liquid)
	return WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       WithdrawTargetLiquid,
		Data:       WithdrawTargetData(data),
	}
}

func (p *WithdrawTarget) LiquidData() (WithdrawTargetLiquidData, error) {
	switch p.Type {

	case WithdrawTargetLiquid:
		var data WithdrawTargetLiquidData
		err := DecodeData(&data, Data(p.Data))
		return data, err

	default:
		return WithdrawTargetLiquidData{}, ErrInvalidDataType
	}
}

// WithdrawTargetLightningData data type for WithdrawTargetType Lightning
type WithdrawTargetLightningData struct {
	WithdrawTargetCryptoData
	Invoice string `json:"invoice,omitempty"`
}

func FromLightningData(withdrawID WithdrawID, lightning WithdrawTargetLightningData) WithdrawTarget {
	if withdrawID > 0 {
		lightning.WithdrawTargetCryptoData.Chain = "lightning"
	}
	data, _ := EncodeData(&lightning)
	return WithdrawTarget{
		WithdrawID: withdrawID,
		Type:       WithdrawTargetLightning,
		Data:       WithdrawTargetData(data),
	}
}

func (p *WithdrawTarget) LightningData() (WithdrawTargetLightningData, error) {
	switch p.Type {

	case WithdrawTargetLightning:
		var data WithdrawTargetLightningData
		err := DecodeData(&data, Data(p.Data))
		return data, err

	default:
		return WithdrawTargetLightningData{}, ErrInvalidDataType
	}
}
