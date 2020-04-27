// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
	"time"
)

func newFloat(value Float) ZeroFloat {
	return &value
}

func refAccountOperation() AccountOperation {
	return AccountOperation{
		ID:               1,
		PrevID:           0,
		AccountID:        3,
		SynchroneousType: SynchroneousTypeSync,
		OperationType:    OperationTypeDeposit,
		ReferenceID:      4,
		Timestamp:        time.Now().UTC().Truncate(time.Second),
		Amount:           newFloat(1.0),
		Balance:          newFloat(10.0),
		LockAmount:       newFloat(2.0),
		TotalLocked:      newFloat(8.0),
	}
}

func amountAccountOperation(amount, balance, lockAmount, totalLocked Float) AccountOperation {
	result := refAccountOperation()
	*result.Amount = amount
	*result.Balance = balance
	*result.LockAmount = lockAmount
	*result.TotalLocked = totalLocked
	return result
}

func TestNewAccountOperation(t *testing.T) {
	t.Parallel()

	ref := refAccountOperation()

	type args struct {
		ID               AccountOperationID
		prevID           AccountOperationID
		accountID        AccountID
		synchroneousType SynchroneousType
		operationType    OperationType
		referenceID      RefID
		timestamp        time.Time
		amount           Float
		balance          Float
		lockAmount       Float
		totalLocked      Float
	}
	tests := []struct {
		name string
		args args
		want AccountOperation
	}{
		{"CheckArgsOrder", args{ref.ID, ref.PrevID, ref.AccountID, ref.SynchroneousType, ref.OperationType, ref.ReferenceID, ref.Timestamp, *ref.Amount, *ref.Balance, *ref.LockAmount, *ref.TotalLocked}, ref},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := NewAccountOperation(tt.args.ID, tt.args.prevID, tt.args.accountID, tt.args.synchroneousType, tt.args.operationType, tt.args.referenceID, tt.args.timestamp, tt.args.amount, tt.args.balance, tt.args.lockAmount, tt.args.totalLocked)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewInitOperation(t *testing.T) {
	t.Parallel()

	type args struct {
		accountID   AccountID
		referenceID RefID
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Init", args{42, 0}, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInitOperation(tt.args.accountID, tt.args.referenceID); got.PreCheck() != tt.want {
				t.Errorf("NewInitOperation() = %v, want %v", got.IsValid(), tt.want)
			}
		})
	}
}

func TestAccountOperation_IsValid(t *testing.T) {
	t.Parallel()

	type fields struct {
		p AccountOperation
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default", fields{AccountOperation{}}, false},
		{"Valid", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 1.0, 1.0)}, true},
		{"Init", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeInit, 0, time.Now(), 0.0, 0.0, 0.0, 0.0)}, true},
		{"NotInit", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeDeposit, 0, time.Now(), 0.0, 0.0, 0.0, 0.0)}, false},
		{"ValidUTC", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now().UTC(), 1.0, 1.0, 1.0, 1.0)}, true},

		{"InvalidID", fields{NewAccountOperation(0, 0, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidPrevID", fields{NewAccountOperation(1, 1, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidAccountID", fields{NewAccountOperation(1, 0, 0, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidSynchroneousType", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeInvalid, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidOperationType", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeInvalid, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},

		// test with amount and lock
		{"ValidNegativeAmount", fields{amountAccountOperation(-1.0, 1.0, 0.0, 1.0)}, true},
		{"ValidNegativeTotalLocked", fields{amountAccountOperation(0.0, 1.0, -1.0, 1.0)}, true},

		{"InvalidTotalLockedAndBalance", fields{amountAccountOperation(0.0, 1.0, 0.0, 2.0)}, false},
		{"InvalidBalance", fields{amountAccountOperation(0.0, -1.0, 0.0, 0.0)}, false},
		{"InvalidLockTotalLocked", fields{amountAccountOperation(0.0, 1.0, 0.0, -1.0)}, false},
		{"InvalidTooManyLocked", fields{amountAccountOperation(0.0, 1.0, 0.0, 2.0)}, false},

		{"InvalidBalanceToLow", fields{amountAccountOperation(2.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidTotalLockedToLow", fields{amountAccountOperation(0.0, 0.0, 2.0, 1.0)}, false},

		// test void operation
		{"InvalidZero", fields{amountAccountOperation(0.0, 0.0, 0.0, 0.0)}, false},
		{"InvalidVoidOperation", fields{amountAccountOperation(0.0, 1.0, 0.0, 1.0)}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields.p
			if got := p.IsValid(); got != tt.want {
				t.Errorf("AccountOperation.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountOperation_PreCheck(t *testing.T) {
	t.Parallel()

	type fields struct {
		p AccountOperation
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default", fields{AccountOperation{}}, false},
		{"Valid", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 1.0, 1.0)}, true},
		{"ValidUTC", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now().UTC(), 1.0, 1.0, 1.0, 1.0)}, true},

		// Valid PreCheck
		{"InvalidID", fields{NewAccountOperation(0, 0, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, true},
		{"InvalidPrevID", fields{NewAccountOperation(1, 1, 42, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, true},

		// Invvalid PreCheck
		{"InvalidAccountID", fields{NewAccountOperation(1, 0, 0, SynchroneousTypeSync, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidSynchroneousType", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeInvalid, OperationTypeNone, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidOperationType", fields{NewAccountOperation(1, 0, 42, SynchroneousTypeSync, OperationTypeInvalid, 0, time.Now(), 1.0, 1.0, 0.0, 0.0)}, false},

		// test with amount and lock
		{"ValidNegativeAmount", fields{amountAccountOperation(-1.0, 1.0, 0.0, 1.0)}, true},
		{"ValidNegativeTotalLocked", fields{amountAccountOperation(0.0, 1.0, -1.0, 1.0)}, true},

		{"InvalidTotalLockedAndBalance", fields{amountAccountOperation(0.0, 1.0, 0.0, 2.0)}, false},
		{"InvalidBalance", fields{amountAccountOperation(0.0, -1.0, 0.0, 0.0)}, false},
		{"InvalidLockTotalLocked", fields{amountAccountOperation(0.0, 1.0, 0.0, -1.0)}, false},
		{"InvalidTooManyLocked", fields{amountAccountOperation(0.0, 1.0, 0.0, 2.0)}, false},

		{"InvalidBalanceToLow", fields{amountAccountOperation(2.0, 1.0, 0.0, 0.0)}, false},
		{"InvalidTotalLockedToLow", fields{amountAccountOperation(0.0, 0.0, 2.0, 1.0)}, false},

		// test void operation
		{"InvalidZero", fields{amountAccountOperation(0.0, 0.0, 0.0, 0.0)}, false},
		{"InvalidVoidOperation", fields{amountAccountOperation(0.0, 1.0, 0.0, 1.0)}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields.p

			if got := p.PreCheck(); got != tt.want {
				t.Errorf("AccountOperation.PreCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
