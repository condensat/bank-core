// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package common

const (
	chanPrefix = "Condensat.Accounting."

	CurrencyInfoSubject         = chanPrefix + "Currency.Info"
	CurrencyCreateSubject       = chanPrefix + "Currency.Create"
	CurrencyListSubject         = chanPrefix + "Currency.List"
	CurrencySetAvailableSubject = chanPrefix + "Currency.SetAvailable"

	AccountCreateSubject    = chanPrefix + "Account.Create"
	AccountInfoSubject      = chanPrefix + "Account.Info"
	AccountListSubject      = chanPrefix + "Account.List"
	AccountHistorySubject   = chanPrefix + "Account.History"
	AccountSetStatusSubject = chanPrefix + "Account.SetStatus"
	AccountOperationSubject = chanPrefix + "Account.Operation"
	AccountTransferSubject  = chanPrefix + "Account.Transfer"

	AccountTransferWithdrawSubject = chanPrefix + "Account.TransferWithdraw"

	BatchWithdrawListSubject = chanPrefix + "BatchWithdraw.List"
)
