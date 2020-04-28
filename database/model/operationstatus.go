// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import "time"

type OperationStatus struct {
	OperationInfoID ID        `gorm:"unique_index;not null"`           // [FK] Reference to OperationInfo table
	LastUpdate      time.Time `gorm:"index;not null;type:timestamp"`   // Last update timestamp
	State           string    `gorm:"index;not null;type:varchar(16)"` // [enum] Operation synchroneous state (open, close)
}
