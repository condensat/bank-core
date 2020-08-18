// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

type Address string
type PubKey string
type BlindingKey string
type Transaction string
type TransactionID string

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

type RawTransaction map[string]interface{}

type RawTransactionBitcoin struct {
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

type RawTransactionLiquid struct {
	Txid     string `json:"txid"`
	Hash     string `json:"hash"`
	Wtxid    string `json:"wtxid"`
	Withash  string `json:"withash"`
	Version  int    `json:"version"`
	Size     int    `json:"size"`
	Vsize    int    `json:"vsize"`
	Weight   int    `json:"weight"`
	Locktime int    `json:"locktime"`
	Vin      []struct {
		Txid      string `json:"txid"`
		Vout      int    `json:"vout"`
		ScriptSig struct {
			Asm string `json:"asm"`
			Hex string `json:"hex"`
		} `json:"scriptSig"`
		IsPegin     bool     `json:"is_pegin"`
		Sequence    int64    `json:"sequence"`
		Txinwitness []string `json:"txinwitness"`
	} `json:"vin"`
	Vout []struct {
		ValueMinimum              float64 `json:"value-minimum,omitempty"`
		ValueMaximum              float64 `json:"value-maximum,omitempty"`
		CtExponent                int     `json:"ct-exponent,omitempty"`
		CtBits                    int     `json:"ct-bits,omitempty"`
		Surjectionproof           string  `json:"surjectionproof,omitempty"`
		Valuecommitment           string  `json:"valuecommitment,omitempty"`
		Assetcommitment           string  `json:"assetcommitment,omitempty"`
		Commitmentnonce           string  `json:"commitmentnonce"`
		CommitmentnonceFullyValid bool    `json:"commitmentnonce_fully_valid"`
		N                         int     `json:"n"`
		ScriptPubKey              struct {
			Asm       string   `json:"asm"`
			Hex       string   `json:"hex"`
			ReqSigs   int      `json:"reqSigs"`
			Type      string   `json:"type"`
			Addresses []string `json:"addresses"`
		} `json:"scriptPubKey,omitempty"`
		Value float64 `json:"value,omitempty"`
		Asset string  `json:"asset,omitempty"`
	} `json:"vout"`
}

type FundRawTransactionOptions struct {
	ChangeAddress          string `json:"changeAddress"`
	IncludeWatching        bool   `json:"includeWatching,omitempty"`
	ChangePosition         int    `json:"changePosition,omitempty"`
	SubtractFeeFromOutputs []int  `json:"subtractFeeFromOutputs,omitempty"`
}

type FundedTransaction struct {
	Changepos int     `json:"changepos"`
	Fee       float64 `json:"fee"`
	Hex       string  `json:"hex"`
}

type SignedTransaction struct {
	Complete bool   `json:"complete"`
	Hex      string `json:"hex"`
}
