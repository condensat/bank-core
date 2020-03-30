// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"encoding/json"
	"io/ioutil"

	"github.com/condensat/bank-core"
	"github.com/condensat/secureid"

	sid "github.com/condensat/secureid"

	"github.com/shengdoushi/base58"
)

type Options struct {
	Seed    string `json:"seed"`
	Context string `json:"context"`
	KeyID   uint   `json:"keyId"`
}

type SecureIDKeys struct {
	keys *sid.Keys
}

func FromFile(filename string) bank.SecureID {
	bytes, _ := ioutil.ReadFile(filename)
	var options Options
	err := json.Unmarshal(bytes, &options)
	if err != nil {
		panic(err)
	}
	return FromOptions(options)
}

func FromOptions(options Options) bank.SecureID {
	seed, err := base58.Decode(options.Seed, base58.BitcoinAlphabet)
	if err != nil {
		panic(err)
	}

	return New(
		secureid.SecureInfo{
			Seed:    seed[:],
			Context: options.Context,
		},
		secureid.KeyID(options.KeyID),
	)
}

func New(info secureid.SecureInfo, keyID secureid.KeyID) bank.SecureID {
	keys := secureid.DefaultKeys(info, keyID)
	if keys == nil {
		panic("Invalid SecureID")
	}
	return &SecureIDKeys{
		keys: keys,
	}
}

func (p *SecureIDKeys) ToSecureID(value secureid.Value) (secureid.SecureID, error) {
	return p.keys.SecureIDFromValue(value)
}

func (p *SecureIDKeys) FromSecureID(secureID secureid.SecureID) (secureid.Value, error) {
	return p.keys.ValueFromSecureID(secureID)
}
