// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"context"
	"sync"

	"github.com/condensat/bank-core/security/utils"

	"github.com/shengdoushi/base58"
)

const (
	KeyPasswordHashSeed = "Security.KeyPasswordHashSeed"
)

type HashSeed struct {
	sync.Mutex
	hashSeed HashSeedKey
}

func PasswordHashSeedContext(ctx context.Context, passwordHashSeed string) context.Context {
	data, err := base58.Decode(passwordHashSeed, DefaultAlphabet)
	defer utils.Memzero(data[:])
	if err != nil {
		panic(err)
	}

	hash := utils.HashBuffers(data)
	defer utils.Memzero(hash[:])

	var hashSeed HashSeedKey
	copy(hashSeed[:], hash[:HashSeedKeySize])
	defer utils.Memzero(hashSeed[:])

	xorHashSeed(ctx, hashSeed)

	seed := base58.Encode(hashSeed[:], DefaultAlphabet)
	return context.WithValue(ctx, KeyPasswordHashSeed, seed)
}

func PasswordHashSalt(ctx context.Context) HashSeedKey {
	passwordHashSeed := ctx.Value(KeyPasswordHashSeed).(string)
	data, err := base58.Decode(passwordHashSeed, DefaultAlphabet)
	if err != nil {
		panic(err)
	}
	defer utils.Memzero(data)

	var hashSeed HashSeedKey
	copy(hashSeed[:], data[:HashSeedKeySize])
	xorHashSeed(ctx, hashSeed)

	return hashSeed
}
