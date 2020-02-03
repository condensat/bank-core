// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package security

import (
	"context"
	"sync"

	"crypto/ed25519"

	"github.com/condensat/bank-core"
	sodium "github.com/condensat/bank-core/security/internal/libsodium"
	"github.com/condensat/bank-core/security/utils"
)

type Key struct {
	sync.Mutex
	secretKey SecretKey
}

func NewSeed() SeedKey {
	var result SeedKey

	err := utils.GenerateRand(result[:])
	if err != nil {
		panic("Failed to generate new seed")
	}

	return result
}

func NewKey(ctx context.Context) *Key {
	seed, err := GenerateSeed()
	if err != nil {
		utils.Memzero(seed[:])
		return nil
	}
	return FromSeed(ctx, seed)
}

func FromSeed(ctx context.Context, seed SeedKey) *Key {
	defer utils.Memzero(seed[:])

	privateKey := ed25519.NewKeyFromSeed(seed[:])
	defer utils.Memzero(privateKey[:])

	// check key conversion
	var pubKey SignaturePublicKey
	defer utils.Memzero(pubKey[:])
	copy(pubKey[:], privateKey[32:])

	publicKey := ed25519.PublicKey(pubKey[:])
	defer utils.Memzero(publicKey[:])

	err := sodium.VerifyKeys(publicKey, privateKey)
	if err != nil {
		return nil
	}

	p := new(Key)

	var secretKey SecretKey
	copy(secretKey[:], privateKey)
	defer utils.Memzero(secretKey[:])

	p.secretKey = xorKey(ctx, secretKey)

	return p
}

func (p *Key) Wipe() {
	defer utils.Memzero(p.secretKey[:])
}

func (p *Key) privateKey(ctx context.Context) SecretKey {
	p.Lock()
	defer p.Unlock()

	return xorKey(ctx, p.secretKey)
}

// Signature

func (p *Key) SignPublicKey(ctx context.Context) SignaturePublicKey {
	var result SignaturePublicKey
	privateKey := p.privateKey(ctx)

	defer utils.Memzero(privateKey[:])

	copy(result[:], privateKey[32:])
	return result
}

func (p *Key) SignMessage(ctx context.Context, message []byte) ([]byte, error) {
	privateKey := p.privateKey(ctx)
	defer utils.Memzero(privateKey[:])

	signatureKey := SignatureSecretKey(privateKey)
	defer utils.Memzero(signatureKey[:])

	return Sign(signatureKey, message)
}

// Encryption

func fromSignaturePublicKey(ctx context.Context, publicKey SignaturePublicKey) EncryptionPublicKey {
	result, err := sodium.ConvertPublicKey(publicKey[:])
	if err != nil {
		panic(err)
	}

	return result
}

func (p *Key) Public(ctx context.Context) EncryptionPublicKey {
	publicKey := p.SignPublicKey(ctx)
	defer utils.Memzero(publicKey[:])

	return fromSignaturePublicKey(ctx, publicKey)
}

func (p *Key) private(ctx context.Context) EncryptionPrivateKey {
	privateKey := p.privateKey(ctx)
	defer utils.Memzero(privateKey[:])
	result, err := sodium.ConvertSecretKey(privateKey[:])
	if err != nil {
		panic(err)
	}

	return result
}

func (p *Key) EncryptFor(ctx context.Context, to EncryptionPublicKey, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, bank.ErrNoData
	}
	defer utils.Memzero(to[:])

	from := p.private(ctx)
	defer utils.Memzero(from[:])

	return EncryptFor(from, to, data)
}

func (p *Key) DecryptFrom(ctx context.Context, from EncryptionPublicKey, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, bank.ErrNoData
	}
	defer utils.Memzero(from[:])

	to := p.private(ctx)
	defer utils.Memzero(to[:])

	return DecryptFrom(from, to, data)
}

// Authentication

func (p *Key) AuthenticationKey(ctx context.Context) AuthenticationKey {
	publicKey := p.Public(ctx)
	defer utils.Memzero(publicKey[:])
	salt := utils.HashString("AuthenticationKey")
	defer utils.Memzero(salt[:])

	var result AuthenticationKey
	digest := utils.HashBuffers(
		salt,
		publicKey[:],
	)
	defer utils.Memzero(digest[:])

	copy(result[:], digest[:AuthenticationKeySize])
	return result
}
