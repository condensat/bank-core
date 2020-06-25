// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

type Address string
type Transaction string

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

	// gettransaction
	Details []struct {
		Address       string  `json:"address"`
		Category      string  `json:"category"`
		Amount        float64 `json:"amount"`
		AmountBlinder string  `json:"amountblinder"`
		Asset         string  `json:"asset"`
		AssetBlinder  string  `json:"assetblinder"`
		Label         string  `json:"label"`
		Vout          int     `json:"vout"`
	}

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

type SpendInfo struct {
	Address string
	Amount  float64
}

type RawTransaction struct {
	Hash     string `json:"hash"`
	Locktime int    `json:"locktime"`
	Size     int    `json:"size"`
	Txid     string `json:"txid"`
	Version  int    `json:"version"`
	Vin      []struct {
		ScriptSig struct {
			Asm string `json:"asm"`
			Hex string `json:"hex"`
		} `json:"scriptSig"`
		Sequence int64  `json:"sequence"`
		Txid     string `json:"txid"`
		Vout     int    `json:"vout"`
	} `json:"vin"`
	Vout []struct {
		N            int `json:"n"`
		ScriptPubKey struct {
			Addresses []string `json:"addresses"`
			Asm       string   `json:"asm"`
			Hex       string   `json:"hex"`
			ReqSigs   int      `json:"reqSigs"`
			Type      string   `json:"type"`
		} `json:"scriptPubKey"`
		Value float64 `json:"value"`
	} `json:"vout"`
	Vsize  int `json:"vsize"`
	Weight int `json:"weight"`
}

type FundedTransaction struct {
	Changepos int     `json:"changepos"`
	Fee       float64 `json:"fee"`
	Hex       string  `json:"hex"`
}
