// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

type Address string

type TransactionInfo struct {
	// Bitcoin
	TxID          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       Address `json:"address"`
	Label         string  `json:"label,omitempty"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Amount        float64 `json:"amount"`
	Confirmations int64   `json:"confirmations"`
	Spendable     bool    `json:"spendable"`
	Solvable      bool    `json:"solvable"`
	Desc          string  `json:"desc,omitempty"`
	Safe          bool    `json:"safe"`

	// Liquid Specific
	AssetCommitment  string `json:"assetcommitment"`
	Asset            string `json:"asset"`
	AmountCommitment string `json:"amountcommitment"`
	AmountBlinder    string `json:"amountblinder"`
	AssetBlinder     string `json:"assetblinder"`
}

type AddressInfo struct {
	Address        string `json:"address"`
	ScriptPubKey   string `json:"scriptPubKey"`
	Ismine         bool   `json:"ismine"`
	Solvable       bool   `json:"solvable"`
	Desc           string `json:"desc"`
	IsWatchonly    bool   `json:"iswatchonly"`
	IsScript       bool   `json:"isscript"`
	IsWitness      bool   `json:"iswitness"`
	WitnessVersion int    `json:"witness_version"`
	WitnessProgram string `json:"witness_program"`
	Script         string `json:"script"`
	Hex            string `json:"hex"`
	Pubkey         string `json:"pubkey"`
	Embedded       struct {
		IsScript       bool   `json:"isscript"`
		IsWitness      bool   `json:"iswitness"`
		WitnessVersion int    `json:"witness_version"`
		WitnessProgram string `json:"witness_program"`
		Pubkey         string `json:"pubkey"`
		Address        string `json:"address"`
		ScriptPubKey   string `json:"scriptPubKey"`
	} `json:"embedded"`
	Label               string `json:"label"`
	IsChange            bool   `json:"ischange"`
	Timestamp           int    `json:"timestamp"`
	HdKeyPath           string `json:"hdkeypath"`
	HdSeedID            string `json:"hdseedid"`
	HdMasterfingerprint string `json:"hdmasterfingerprint"`
	Labels              []struct {
		Name    string `json:"name"`
		Purpose string `json:"purpose"`
	} `json:"labels"`

	// Liquid Specific
	Confidential    string `json:"confidential"`
	ConfidentialKey string `json:"confidential_key"`
	Unconfidential  string `json:"unconfidential"`
}

type UTXOInfo struct {
	TxID string `json:"txid"`
	Vout int    `json:"vout"`
}
