// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type BatchWithdraw struct {
	BatchID    BatchID    `gorm:"unique_index:idx_batch_withdraw;index;not null"`                     // [FK] Reference to Batch table
	WithdrawID WithdrawID `gorm:"unique_index:idx_withdraw;unique_index:idx_batch_withdraw;not null"` // [FK] Reference to Withdraw table
}
