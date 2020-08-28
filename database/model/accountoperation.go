// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"math"
	"time"
)

type AccountOperationID ID

// AccountOperation model
type AccountOperation struct {
	ID        AccountOperationID `gorm:"primary_key;unique_index:idx_id_previd;"` // [PK] AccountOperation
	AccountID AccountID          `gorm:"index;not null"`                          // [FK] Reference to Account table

	SynchroneousType SynchroneousType `gorm:"index;not null;type:varchar(16)"` // [enum] Operation synchroneous type (sync, async-start, async-end)
	OperationType    OperationType    `gorm:"index;not null;type:varchar(16)"` // [enum] Determine table for ReferenceID (deposit, withdraw, transfer, adjustment, none, other)
	ReferenceID      RefID            `gorm:"index;not null"`                  // [optional - FK] Reference to related table with OperationType

	Timestamp time.Time `gorm:"index;not null;type:timestamp"` // Operation timestamp
	Amount    ZeroFloat `gorm:"default:0;not null"`            // Operation amount (can be negative)
	Balance   ZeroFloat `gorm:"default:0;not null"`            // Account balance (strictly positive or zero)

	LockAmount  ZeroFloat `gorm:"default:0;not null"` // Operation amount (can be negative)
	TotalLocked ZeroFloat `gorm:"default:0;not null"` // Total locked (strictly positive or zero and less or equal than Balance)
}

func NewAccountOperation(ID AccountOperationID, accountID AccountID, synchroneousType SynchroneousType, operationType OperationType, referenceID RefID, timestamp time.Time, amount, balance, lockAmount, totalLocked Float) AccountOperation {
	return AccountOperation{
		ID:        ID,
		AccountID: accountID,

		SynchroneousType: synchroneousType,
		OperationType:    operationType,
		ReferenceID:      referenceID,

		Timestamp: timestamp.UTC().Truncate(time.Second),
		Amount:    &amount,
		Balance:   &balance,

		LockAmount:  &lockAmount,
		TotalLocked: &totalLocked,
	}
}

func NewInitOperation(accountID AccountID, referenceID RefID) AccountOperation {
	return NewAccountOperation(0,
		accountID,
		SynchroneousTypeSync,
		OperationTypeInit,
		referenceID,
		time.Now(),
		0.0, 0.0,
		0.0, 0.0,
	)
}

func (p *AccountOperation) IsValid() bool {
	return p.ID > 0 &&
		p.AccountID > 0 &&

		// check enums
		p.OperationType.Valid() &&
		p.SynchroneousType.Valid() &&

		// check pointers
		p.Amount != nil && p.Balance != nil &&
		p.LockAmount != nil && p.TotalLocked != nil &&

		// check positive balance and TotalLocked
		*p.Balance >= 0 &&
		*p.TotalLocked >= 0 &&
		// check for max total locked
		*p.TotalLocked <= *p.Balance &&

		// check amount less or equals than balance
		*p.Amount <= *p.Balance &&
		// check lockAmount less or equals than totalLocked
		*p.LockAmount <= *p.TotalLocked &&

		// Check for zero operation
		// allow zero operation for OperationTypeInit
		(p.OperationType == OperationTypeInit || (math.Abs(float64(*p.Amount)) > 0.0 || math.Abs(float64(*p.LockAmount)) > 0.0))
}

func (p *AccountOperation) PreCheck() bool {
	// deepcopy
	operation := *p
	// overwite operation IDs
	operation.ID = 1

	// operation should be valid
	return operation.IsValid()
}
