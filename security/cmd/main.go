// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/condensat/bank-core/security"
	"github.com/condensat/bank-core/security/utils"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, security.KeyPrivateKeySalt, utils.GenerateRandN(32))

	message := []byte("Hello, Box!")

	for i := 0; i < 1; i++ {
		testKeyEncoding(ctx)
	}

	for i := 0; i < 1; i++ {
		testSignature(ctx, message)
		testAuthenticate(ctx, message)

		decrypted := testEncryption(ctx, message)
		if !bytes.Equal(decrypted, message[:]) {
			panic("Wrong decrypted message")
		}
		fmt.Println(string(decrypted))
	}
}

func vanitySeed(ctx context.Context, prefix string) security.SeedKey {
	for {
		seed := security.EncodeSeedKey(security.NewSeed())
		if strings.HasPrefix(seed, prefix) {
			key, _ := security.DecodeSeedKey(seed)
			return key
		}
	}
}

func vanityPublic(ctx context.Context, prefix string) *security.Key {
	for {
		k := security.NewKey(ctx)
		pub := security.EncodePublicKey(k.Public(ctx))

		if strings.HasPrefix(pub, prefix) {
			return k
		}
	}
}

func testKeyEncoding(ctx context.Context) {
	u := security.FromSeed(ctx, security.NewSeed())
	if u == nil {
		panic("Unable to create keys")
	}
	defer u.Wipe()

	{
		seed := vanitySeed(ctx, "🥩")
		defer utils.Memzero(seed[:])

		seedEnc := security.EncodeSeedKey(seed)
		seedDec, err := security.DecodeSeedKey(seedEnc)
		defer utils.Memzero(seedDec[:])
		if err != nil {
			panic("Unable to decode key")
		}
		if !bytes.Equal(seed[:], seedDec[:]) {
			panic("Wrong seed")
		}
		fmt.Println("Valid seed\t: ", seedEnc)
	}
	{
		u := vanityPublic(ctx, "🍙")
		pub := u.Public(ctx)
		defer utils.Memzero(pub[:])

		keyEnc := security.EncodePublicKey(pub)
		keyDec, err := security.DecodePublicKey(keyEnc)
		defer utils.Memzero(keyDec[:])
		if err != nil {
			panic("Unable to decode key")
		}
		if !bytes.Equal(pub[:], keyDec[:]) {
			panic("Wrong key")
		}
		fmt.Println("Valid pub\t: ", keyEnc)
	}
	{
		sigKey := u.SignPublicKey(ctx)
		defer utils.Memzero(sigKey[:])

		keyEnc := security.EncodeSignatureKey(sigKey)
		keyDec, err := security.DecodeSignatureKey(keyEnc)
		defer utils.Memzero(keyDec[:])

		if err != nil {
			panic("Unable to decode key")
		}
		if !bytes.Equal(sigKey[:], keyDec[:]) {
			panic("Wrong key")
		}
		fmt.Println("Valid sign\t: ", keyEnc)
	}
	{
		authKey := u.AuthenticationKey(ctx)
		defer utils.Memzero(authKey[:])

		keyEnc := security.EncodeAuthenticationKey(authKey)
		keyDec, err := security.DecodeAuthenticationKey(keyEnc)
		defer utils.Memzero(keyDec[:])
		if err != nil {
			panic("Unable to decode key")
		}
		if !bytes.Equal(authKey[:], keyDec[:]) {
			panic("Wrong key")
		}
		fmt.Println("Valid auth\t: ", keyEnc)
	}
	{
		seed := security.NewSeed()
		defer utils.Memzero(seed[:])
		auth := security.AuthenticateMessage(u.AuthenticationKey(ctx), seed[:])
		if !security.VerifyMessageAuthentication(u.AuthenticationKey(ctx), auth, seed[:]) {
			panic("VerifyMessageAuthentication failed")
		}

		authtEnc := security.EncodeAuthenticationDigest(auth)
		authDec, err := security.DecodeAuthenticationDigest(authtEnc)
		defer utils.Memzero(authDec[:])
		if err != nil {
			panic("Unable to decode digest")
		}
		if !bytes.Equal(auth[:], authDec[:]) {
			panic("Wrong digest")
		}
		fmt.Println("Valid digest\t: ", authtEnc)
	}
	fmt.Println("-----")
}

func testSignature(ctx context.Context, message []byte) {
	u := security.NewKey(ctx)
	if u == nil {
		panic("Unable to create keys")
	}
	defer u.Wipe()

	signedMessage, err := u.SignMessage(ctx, message)
	if err != nil {
		panic(err)
	}

	ok, err := security.VerifySignature(u.SignPublicKey(ctx), signedMessage)
	if err != nil {
		panic(err)
	}

	if !ok {
		panic("Message not verified")
	}
	fmt.Println("Message verified")
}

func testAuthenticate(ctx context.Context, message []byte) {
	u := security.NewKey(ctx)
	if u == nil {
		panic("Unable to create keys")
	}
	defer u.Wipe()

	auth := security.AuthenticateMessage(u.AuthenticationKey(ctx), message)
	if !security.VerifyMessageAuthentication(u.AuthenticationKey(ctx), auth, message) {
		panic("Message not authenticated")
	}

	fmt.Println("Message authenticated")
}

func testEncryption(ctx context.Context, message []byte) []byte {
	sender := security.NewKey(ctx)
	reciever := security.NewKey(ctx)
	if sender == nil || reciever == nil {
		panic("Unable to create users")
	}
	defer sender.Wipe()
	defer reciever.Wipe()

	data, err := sender.EncryptFor(ctx, reciever.Public(ctx), message)
	if err != nil {
		panic(err)
	}

	decrypted, err := reciever.DecryptFrom(ctx, sender.Public(ctx), data)
	if err != nil {
		panic(err)
	}

	return decrypted
}
