// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
)

type BankStatus struct {
	Users      UsersStatus      `json:"users"`
	Accounting AccountingStatus `json:"accounting"`
	Transfer   TransferStatus   `json:"transfer"`
	Crypto     CryptoStatus     `json:"crypto"`
}

func FetchBankStatus(ctx context.Context) (BankStatus, error) {
	userStatus, err := FetchUserStatus(ctx)
	if err != nil {
		return BankStatus{}, err
	}

	accountingStatus, err := FetchAccountingStatus(ctx)
	if err != nil {
		return BankStatus{}, err
	}

	transfertStatus, err := FetchTransferStatus(ctx)
	if err != nil {
		return BankStatus{}, err
	}

	cryptoStatus, err := FetchCryptoStatus(ctx)
	if err != nil {
		return BankStatus{}, err
	}

	return BankStatus{
		Users:      userStatus,
		Accounting: accountingStatus,
		Transfer:   transfertStatus,
		Crypto:     cryptoStatus,
	}, nil
}
