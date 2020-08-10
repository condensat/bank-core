// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"

	"github.com/condensat/bank-core/wallet/common"
)

type WalletOptions struct {
	FileName string
	Mode     string
}

func loadOptionsFromFile(fileName string, options interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, options)
	if err != nil {
		return err
	}

	return nil
}

func loadChainsOptionsFromFile(fileName string) ChainsOptions {
	var result ChainsOptions

	err := loadOptionsFromFile(fileName, &result)
	if err != nil {
		return ChainsOptions{}
	}

	return result
}

type ChainOption struct {
	Chain    string `json:"chain"`
	HostName string `json:"hostname"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
}

type ChainsOptions struct {
	Chains []ChainOption `json:"chains"`
}

func (p *ChainsOptions) Names() []string {
	var result []string
	for _, option := range p.Chains {
		result = append(result, option.Chain)
	}
	return result
}

type SsmOption struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

type SsmChain struct {
	Device      string `json:"device"`
	Chain       string `json:"chain"`
	Fingerprint string `json:"fingerprint"`
}

type SsmOptions struct {
	Ssm struct {
		Devices []SsmOption `json:"devices"`
		Chains  []SsmChain  `json:"chains"`
	} `json:"ssm"`

	TorProxy string `json:"tor_proxy"`
}

func (p *SsmOptions) Devices() []string {
	var result []string
	for _, option := range p.Ssm.Devices {
		result = append(result, option.Name)
	}
	return result
}

func loadSsmOptionsFromFile(fileName string) SsmOptions {
	var result SsmOptions

	err := loadOptionsFromFile(fileName, &result)
	if err != nil {
		return SsmOptions{}
	}

	return result
}

func OptionArgs(args *WalletOptions) {
	if args == nil {
		panic("Invalid wallet options")
	}

	flag.StringVar(&args.FileName, "chains", "chains.json", "Json file for (default chain.json)")
	cryptoMode := string(common.CryptoModeBitcoinCore)
	flag.StringVar(&args.Mode, "cryptoMode", cryptoMode, "Crypto mode for new address & signature (default bitcoin-core)")
}

func parseCryptoMode(cryptoMode string) common.CryptoMode {
	result := common.CryptoMode(cryptoMode)
	switch common.CryptoMode(cryptoMode) {
	case common.CryptoModeCryptoSsm:
		return result

	default:
		return common.CryptoModeBitcoinCore
	}
}
