// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"crypto/sha512"
	"fmt"

	"github.com/condensat/bank-core/security/secureid"
)

func main() {
	secret := sha512.Sum512([]byte("secret_seed"))

	// test with keys
	fmt.Println("--- Test with keys")
	keys := secureid.NewKeys(secureid.SecureInfo{
		Seed:    secureid.Seed(secret[:]),
		Context: "testSecureID",
	}, secureid.Version, 0)

	testSecureIDKeys(keys, 42)
	testSecureIDKeys(keys, 43)

	// test with info
	fmt.Println("--- Test with info")
	info := secureid.SecureInfo{
		Seed:    secureid.Seed(secret[:]),
		Context: "testSecureIDInfo",
	}
	testSecureIDInfo(info, secureid.Version0, 0, 100)
	testSecureIDInfo(info, secureid.Version1, 1, 100)
	testSecureIDInfo(info, secureid.Version, 2, 101)

	// test keys and info
	info = secureid.SecureInfo{
		Seed:    secureid.Seed(secret[:]),
		Context: "loop",
	}
	count := 10
	{
		keys = secureid.NewKeys(info, secureid.Version, 3)

		fmt.Println("--- Test in loop with keys")
		for i := 0; i < count; i++ {
			id := secureid.Value(i + 1)
			testSecureIDKeys(keys, id)
		}
	}
	{
		fmt.Println("--- Test in loop with info")
		for i := 0; i < count; i++ {
			id := secureid.Value(i + 1)
			testSecureIDInfo(info, secureid.Version, 4, id)
		}
	}
}

func testSecureIDKeys(keys *secureid.Keys, value secureid.Value) {
	secureID, err := keys.SecureIDFromValue(value)
	if err != nil {
		panic(err)
	}

	val, err := keys.ValueFromSecureID(secureID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", toString(secureID))
	checkValues(value, val)
}

func testSecureIDInfo(info secureid.SecureInfo, version secureid.ProtocolVersion, keyID secureid.KeyID, value secureid.Value) {
	secureID, err := secureid.SecureIDFromValue(info, version, keyID, value)
	if err != nil {
		panic(err)
	}

	val, err := secureID.Value(info, secureID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", toString(secureID))
	checkValues(value, val)

}

func checkValues(ref, value secureid.Value) {
	if value != ref {
		panic(value)
	}
}

func toString(secureID secureid.SecureID) string {
	// data, _ := json.Marshal(&secureID)
	// return string(data)
	return fmt.Sprintf("%v", secureID)
}
