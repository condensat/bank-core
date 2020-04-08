// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"errors"
)

type AccountStatus String

const (
	AccountStatusInvalid AccountStatus = ""

	AccountStatusCreated  AccountStatus = "created"
	AccountStatusNormal   AccountStatus = "normal"
	AccountStatusLocked   AccountStatus = "locked"
	AccountStatusDisabled AccountStatus = "disabled"
)

var (
	ErrAccountStatusInvalid = errors.New("Invalid AccountStatus")
)

type AccountState struct {
	AccountID AccountID     `gorm:"unique_index;not null"`           // [FK] Reference to Account table
	State     AccountStatus `gorm:"index;not null;type:varchar(16)"` // AccountStatus [normal, locked, disabled]
}

func (p AccountStatus) Valid() bool {
	switch p {
	case AccountStatusCreated:
		fallthrough
	case AccountStatusNormal:
		fallthrough
	case AccountStatusLocked:
		fallthrough
	case AccountStatusDisabled:
		return true

	default:
		return false
	}
}

func ParseAccountStatus(str string) AccountStatus {
	ret := AccountStatus(str)
	if !ret.Valid() {
		return AccountStatusInvalid
	}
	return ret
}

func (p AccountStatus) String() string {
	if !p.Valid() {
		return string(AccountStatusInvalid)
	}
	return string(p)
}

func knownAccountStatus() []AccountStatus {
	return []AccountStatus{
		AccountStatusInvalid,

		AccountStatusCreated,
		AccountStatusNormal,
		AccountStatusLocked,
		AccountStatusDisabled,
	}
}
