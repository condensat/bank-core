// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type User struct {
	ID    uint64 `gorm:"primary_key"`
	Name  string `gorm:"size:64;unique;not null"`
	Email string `gorm:"size:256;unique;not null"`
}
