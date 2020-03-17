// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type Credential struct {
	UserID       uint64 `gorm:"unique_index"`
	LoginHash    string `gorm:"size:64;not null;index"`
	PasswordHash string `gorm:"size:64;not null;index"`
	TOTPSecret   string `gorm:"size:64;not null"`
}
