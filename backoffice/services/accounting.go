// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/logger"

	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"

	"github.com/condensat/secureid"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
)

type CurrencyBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Locked   float64 `json:"locked"`
}

type AccountingStatus struct {
	Count    int               `json:"count"`
	Active   int               `json:"active"`
	Balances []CurrencyBalance `json:"balances"`
}

func FetchAccountingStatus(ctx context.Context) (AccountingStatus, error) {
	db := appcontext.Database(ctx)

	accountsInfo, err := database.AccountsInfos(db)
	if err != nil {
		return AccountingStatus{}, err
	}

	var balances []CurrencyBalance
	for _, account := range accountsInfo.Accounts {
		balances = append(balances, CurrencyBalance{
			Currency: account.CurrencyName,
			Balance:  account.Balance,
			Locked:   account.TotalLocked,
		})
	}

	return AccountingStatus{
		Count:    accountsInfo.Count,
		Active:   accountsInfo.Active,
		Balances: balances,
	}, nil
}

// UserAccountListRequest holds args for useraccountlist requests
type UserAccountListRequest struct {
	apiservice.SessionArgs
	UserID string `json:"userId"`
}

// UserAccountListResponse holds response for useraccountlist request
type UserAccountListResponse struct {
	UserID     string           `json:"userId"`
	Accounts   []string         `json:"accounts"`
	Accounting AccountingStatus `json:"accounting"`
}

func (p *DashboardService) UserAccountList(r *http.Request, request *UserAccountListRequest, reply *UserAccountListResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.UserAccountListRequest")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "UserAccountListRequest")

	// Get userID from session
	request.SessionID = apiservice.GetSessionCookie(r)
	sessionID := sessions.SessionID(request.SessionID)

	isAdmin, log, err := isUserAdmin(ctx, log, sessionID)
	if err != nil {
		log.WithError(err).
			WithField("RoleName", model.RoleNameAdmin).
			Error("UserHasRole failed")
		return ErrPermissionDenied
	}
	if !isAdmin {
		log.WithError(err).
			Error("User is not Admin")
		return ErrPermissionDenied
	}

	sID := appcontext.SecureID(ctx)

	userID, err := sID.FromSecureID("user", sID.Parse(request.UserID))
	if err != nil {
		log.WithError(err).
			WithField("UserID", request.UserID).
			Error("userID FromSecureID failed")
		return ErrPermissionDenied
	}

	var user model.User
	var accounts []string
	var accounting AccountingStatus
	err = db.Transaction(func(db bank.Database) error {
		var err error

		user, err = database.FindUserById(db, model.UserID(userID))
		if err != nil {
			return err
		}

		accountsInfo, err := database.AccountsInfosByUser(db, model.UserID(userID))
		if err != nil {
			return err
		}

		var balances []CurrencyBalance
		for _, account := range accountsInfo.Accounts {
			balances = append(balances, CurrencyBalance{
				Currency: account.CurrencyName,
				Balance:  account.Balance,
				Locked:   account.TotalLocked,
			})
		}

		accounting = AccountingStatus{
			Count:    accountsInfo.Count,
			Active:   accountsInfo.Active,
			Balances: balances,
		}

		accountIDs, err := database.GetUserAccounts(db, user.ID)
		if err != nil {
			return err
		}
		for _, accountID := range accountIDs {
			secureID, err := sID.ToSecureID("account", secureid.Value(uint64(accountID)))
			if err != nil {
				return err
			}

			accounts = append(accounts, sID.ToString(secureID))
		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("UserAccountList failed")
		return apiservice.ErrServiceInternalError
	}

	*reply = UserAccountListResponse{
		UserID:     request.UserID,
		Accounts:   accounts,
		Accounting: accounting,
	}

	return nil
}
