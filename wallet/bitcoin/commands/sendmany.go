// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/condensat/bank-core/wallet/rpc"
)

type EstimateMode string

const (
	SpendMinConf     = 3
	SpendReplaceable = false

	EstimateModeUnest        = "UNSET"
	EstimateModeEconomical   = "ECONOMICAL"
	EstimateModeConservative = "CONSERVATIVE"
)

var (
	ErrInvalidPayAddress = errors.New("Invalid Pay Address")
	ErrInvalidPayAmount  = errors.New("Invalid Pay Amount")
	ErrInvalidRecipents  = errors.New("Invalid Recipents")
)

type TxID string
type PayInfo struct {
	Address Address
	Amount  float64
}

func SendMany(ctx context.Context, rpcClient RpcClient, comment string, payinfo []PayInfo) (TxID, error) {
	return SendManyWithFees(ctx, rpcClient, comment, payinfo, nil, EstimateModeEconomical)
}

func SendManyWithFees(ctx context.Context, rpcClient RpcClient, comment string, payinfo []PayInfo, fees []Address, estimateMode EstimateMode) (TxID, error) {
	const dummy = ""
	const confTarget = 1
	amounts := make(map[string]float64)
	for _, info := range payinfo {
		log.Printf("%+v", info)
		if len(info.Address) == 0 {
			return "", ErrInvalidPayAddress
		}
		if info.Amount <= 0.0 {
			return "", ErrInvalidPayAmount
		}
		amounts[string(info.Address)] = info.Amount
	}
	if len(amounts) == 0 {
		return "", ErrInvalidRecipents
	}
	subtractfeefrom := make([]string, 0)
	for _, fee := range fees {
		subtractfeefrom = append(subtractfeefrom, string(fee))
	}

	amountData, err := json.Marshal(amounts)
	if err != nil {
		return "", err
	}
	amountMap := make(map[string]float64)
	err = json.Unmarshal(amountData, &amountMap)
	if err != nil {
		return "", err
	}
	log.Printf("%+v", amountMap)

	var result TxID
	err = callCommand(rpcClient, CmdSendMany, &result, dummy, amountMap, SpendMinConf, comment, subtractfeefrom[:], SpendReplaceable, confTarget, estimateMode)
	if err != nil {
		return "", rpc.ErrRpcError
	}

	return result, nil
}
