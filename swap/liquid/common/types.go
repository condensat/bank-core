// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/utils"
)

const (
	AssetIDLength = 64

	DefaultFeeRate = 150 / 100000000.0 // BTC/Kb
	MinumumFeeRate = 150 / 100000000.0 // BTC/Kb

	AmountPrecision       = 8
	AmountPrecisionFormat = "%.8f"
)

var (
	ErrInvalidAddress  = errors.New("Invalid Address")
	ErrInvalidProposal = errors.New("Invalid Proposal")
	ErrInvalidPayload  = errors.New("Invalid PayLoad")
)

type ConfidentialAddress string
type AssetID string
type Payload string

type ProposalInfo struct {
	ProposerAsset  AssetID
	ProposerAmount float64
	ReceiverAsset  AssetID
	ReceiverAmount float64
}

type SwapProposal struct {
	Timestamp time.Time
	SwapID    uint64
	Address   ConfidentialAddress
	Proposal  ProposalInfo
	FeeRate   float64
	Payload   Payload
}

func (p *ProposalInfo) Args() []string {
	proposerAsset := string(p.ProposerAsset)
	receiverAsset := string(p.ReceiverAsset)

	proposerAmount := fmt.Sprintf(AmountPrecisionFormat, utils.ToFixed(p.ProposerAmount, AmountPrecision))
	receiverAmount := fmt.Sprintf(AmountPrecisionFormat, utils.ToFixed(p.ReceiverAmount, AmountPrecision))

	return []string{
		proposerAsset,
		proposerAmount,
		receiverAsset,
		receiverAmount,
	}
}

func (p *ProposalInfo) Valid() bool {
	return len(p.ProposerAsset) == AssetIDLength &&
		len(p.ReceiverAsset) == AssetIDLength &&
		p.ProposerAmount > 0.0 &&
		p.ReceiverAmount > 0.0
}

func (p *SwapProposal) Encode() ([]byte, error) {
	return messaging.EncodeObject(p)
}

func (p *SwapProposal) Decode(data []byte) error {
	return messaging.DecodeObject(data, messaging.BankObject(p))
}

func (payload Payload) Stdin() io.Reader {
	if len(payload) == 0 {
		return nil
	}
	return strings.NewReader(string(payload))
}

func (payload Payload) Valid() bool {
	if len(payload) == 0 {
		return false
	}
	// check if base64 payload
	decoded, err := base64.StdEncoding.DecodeString(string(payload))
	if err != nil {
		// check if payload is a raw json object
		return isJson([]byte(payload))
	}

	// check if decoded payload is a json object
	return isJson([]byte(decoded))
}

func isJson(data []byte) bool {
	var obj map[string]interface{}
	err := json.Unmarshal(data, &obj)
	return err == nil
}
