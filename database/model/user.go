// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type UserID ID
type UserName String
type UserEmail String

type User struct {
	ID    UserID    `gorm:"primary_key"`
	Name  UserName  `gorm:"size:64;unique;not null"`
	Email UserEmail `gorm:"size:256;unique;not null"`
}
