// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

import (
	"github.com/condensat/bank-core/database/model"
)

func ConvertEntryToOperation(entry AccountEntry) model.AccountOperation {
	amount := model.Float(entry.Amount)
	lockAmount := model.Float(entry.LockAmount)

	// Balance & totalLocked ar computed by database later, must be valid for pre-check
	var balance model.Float
	if balance < amount {
		balance = amount
	}
	var totalLocked model.Float
	if totalLocked < lockAmount {
		totalLocked = lockAmount
	}

	return model.AccountOperation{
		AccountID:        model.AccountID(entry.AccountID),
		SynchroneousType: model.ParseSynchroneousType(entry.SynchroneousType),
		OperationType:    model.ParseOperationType(entry.OperationType),
		ReferenceID:      model.RefID(entry.ReferenceID),

		Amount:  &amount,
		Balance: &balance,

		LockAmount:  &lockAmount,
		TotalLocked: &totalLocked,

		Timestamp: entry.Timestamp,
	}
}

func ConvertOperationToEntry(op model.AccountOperation, label string) AccountEntry {
	return AccountEntry{
		OperationID: uint64(op.ID),
		AccountID:   uint64(op.AccountID),
		ReferenceID: uint64(op.ReferenceID),

		OperationType:    string(op.OperationType),
		SynchroneousType: string(op.SynchroneousType),

		Timestamp: op.Timestamp,
		Label:     label,
		Amount:    float64(*op.Amount),
		Balance:   float64(*op.Balance),

		LockAmount:  float64(*op.LockAmount),
		TotalLocked: float64(*op.TotalLocked),
	}
}
