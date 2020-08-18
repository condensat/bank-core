// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/condensat/bank-core/utils"
)

var (
	ErrInputsError = errors.New("Inputs errors")
)

func SignTx(ctx context.Context, rpcClient RpcClient, chain, inputransaction string, inputs ...SignTxInputs) (SignTxResponse, error) {
	if rpcClient == nil {
		return SignTxResponse{}, ErrInvalidRPCClient
	}

	if len(inputs) == 0 {
		return SignTxResponse{}, ErrInputsError
	}

	var fingerprints string
	var paths string
	var amounts string
	for _, input := range inputs {
		fingerprints = fmt.Sprintf("%s %s", fingerprints, input.Fingerprint)
		paths = fmt.Sprintf("%s %s", paths, input.Path)
		if len(input.ValueCommitment) == 0 {
			amounts = fmt.Sprintf("%s %.8f", amounts, utils.ToFixed(input.Amount, 8))
		} else {
			amounts = fmt.Sprintf("%s %s", amounts, input.ValueCommitment)
		}
	}
	fingerprints = strings.Trim(fingerprints, " ")
	paths = strings.Trim(paths, " ")
	amounts = strings.Trim(amounts, " ")

	var signedTx SignTxResponse
	err := callCommand(rpcClient, CmdSignTx, &signedTx, chain, inputransaction, fingerprints, paths, amounts)
	if err != nil {
		return SignTxResponse{}, err
	}

	return signedTx, nil
}
