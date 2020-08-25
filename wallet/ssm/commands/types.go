// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

type NewAddressResponse struct {
	Chain       string `json:"chain"`
	Address     string `json:"address"`
	PubKey      string `json:"pubkey"`
	BlindingKey string `json:"blinding_key"`
}

type SsmPath struct {
	Chain       string
	Fingerprint string
	Path        string
}

type SignTxInputs struct {
	SsmPath
	Amount          float64
	ValueCommitment string
}

type SignTxResponse struct {
	Chain    string `json:"chain"`
	SignedTx string `json:"signed_tx"`
	Debug    string `json:"debug"`
}
