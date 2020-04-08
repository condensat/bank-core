// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package secureid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/condensat/bank-core"

	"github.com/condensat/secureid"

	"github.com/shengdoushi/base58"
)

var (
	ErrInvalidKeys        = errors.New("Invalid SecureID Keys")
	InvalidSecureIDString = errors.New("Invalid SecureID String")
)

type Options struct {
	Seed    string `json:"seed"`
	Context string `json:"context"`
	KeyID   uint   `json:"keyId"`
}

type KeysMap map[string]*secureid.Keys

type SecureIDKeys struct {
	sync.Mutex
	info  secureid.SecureInfo
	keyID secureid.KeyID

	subKeys KeysMap
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
		info:  info,
		keyID: keyID,

		subKeys: make(KeysMap),
	}
}

func (p *SecureIDKeys) SubKey(context string) *secureid.Keys {
	p.Lock()
	defer p.Unlock()

	if len(context) == 0 {
		context = "default"
	}

	keys, ok := p.subKeys[context]
	if !ok {
		// create sub key with context if not found
		info := secureid.SecureInfo{
			Seed:    p.info.Seed,
			Context: fmt.Sprintf("%s:%s", p.info.Context, context),
		}
		// return new keys
		keys = secureid.DefaultKeys(info, p.keyID)
		p.subKeys[context] = keys
	}

	return keys
}

func (p *SecureIDKeys) ToSecureID(context string, value secureid.Value) (secureid.SecureID, error) {
	keys := p.SubKey(context)
	if keys == nil {
		return secureid.SecureID{}, ErrInvalidKeys
	}
	return keys.SecureIDFromValue(value)
}

func (p *SecureIDKeys) FromSecureID(context string, secureID secureid.SecureID) (secureid.Value, error) {
	keys := p.SubKey(context)
	if keys == nil {
		return secureid.Value(0), ErrInvalidKeys
	}
	return keys.ValueFromSecureID(secureID)
}

func (p *SecureIDKeys) ToString(secureID secureid.SecureID) string {
	if len(secureID.Version) == 0 || len(secureID.Data) == 0 || len(secureID.Check) == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%s:%s", secureID.Version, secureID.Data, secureID.Check)
}

func (p *SecureIDKeys) Parse(secureID string) secureid.SecureID {
	toks := strings.Split(secureID, ":")
	if len(toks) != 3 {
		return secureid.SecureID{}
	}

	return secureid.SecureID{
		Version: toks[0],
		Data:    toks[1],
		Check:   toks[2],
	}
}
