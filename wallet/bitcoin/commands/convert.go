// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
)

func ConvertToRawTransactionBitcoin(tx RawTransaction) (RawTransactionBitcoin, error) {
	var result RawTransactionBitcoin

	err := convertRawTransaction(tx, &result)
	if err != nil {
		return RawTransactionBitcoin{}, err
	}

	return result, nil
}

func ConvertToRawTransactionLiquid(tx RawTransaction) (RawTransactionLiquid, error) {
	var result RawTransactionLiquid

	err := convertRawTransaction(tx, &result)
	if err != nil {
		return RawTransactionLiquid{}, err
	}

	return result, nil
}

func convertRawTransaction(tx RawTransaction, result interface{}) error {
	data, err := json.Marshal(&tx)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		return err
	}

	return nil
}
