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
	ID        AccountOperationID `gorm:"primary_key"`    // [PK] AccountOperation
	PrevID    AccountOperationID `gorm:"index;not null"` // [FK] Reference to previous AccountOperation (0 mean first operation)
	AccountID AccountID          `gorm:"index;not null"` // [FK] Reference to Account table

	SynchroneousType SynchroneousType `gorm:"index;not null;type:varchar(16)"` // [enum] Operation synchroneous type (sync, async-start, async-end)
	OperationType    OperationType    `gorm:"index;not null;type:varchar(16)"` // [enum] Determine table for ReferenceID (deposit, withdraw, transfert, adjustment, none, other)
	ReferenceID      RefID            `gorm:"index;not null"`                  // [optional - FK] Reference to related table with OperationType

	Timestamp time.Time `gorm:"index;not null;type:timestamp"` // Operation timestamp
	Amount    ZeroFloat `gorm:"default:0;not null"`            // Operation amount (can be negative)
	Balance   ZeroFloat `gorm:"default:0;not null"`            // Account balance (strictly positive or zero)

	LockAmount  ZeroFloat `gorm:"default:0;not null"` // Operation amount (can be negative)
	TotalLocked ZeroFloat `gorm:"default:0;not null"` // Total locked (strictly positive or zero and less or equal than Balance)
}

func NewAccountOperation(ID, prevID AccountOperationID, accountID AccountID, synchroneousType SynchroneousType, operationType OperationType, referenceID RefID, timestamp time.Time, amount, balance, lockAmount, totalLocked Float) AccountOperation {
	return AccountOperation{
		ID:        ID,
		PrevID:    prevID,
		AccountID: accountID,

		SynchroneousType: synchroneousType,
		OperationType:    operationType,
		ReferenceID:      referenceID,

		Timestamp: timestamp.UTC(),
		Amount:    &amount,
		Balance:   &balance,

		LockAmount:  &lockAmount,
		TotalLocked: &totalLocked,
	}
}

func (p *AccountOperation) IsValid() bool {
	return p.ID > 0 &&
		p.ID > p.PrevID &&
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

		// Check for void operation
		(math.Abs(float64(*p.Amount)) > 0.0 || math.Abs(float64(*p.LockAmount)) > 0.0)
}

func (p *AccountOperation) PreCheck() bool {
	// deepcopy
	operation := *p
	// overwite operation IDs
	operation.ID = 2
	operation.PrevID = 1

	// operation should be valid
	return operation.IsValid()
}
