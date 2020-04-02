// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"fmt"
	"net/http"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/security/utils"

	"github.com/condensat/bank-core/accounting/client"
	"github.com/condensat/bank-core/api/sessions"

	"github.com/sirupsen/logrus"
)

type AccountingService int

// AccountRequest holds args for accounting requests
type AccountRequest struct {
	SessionArgs
}

// AccountInfo holds account information
type AccountInfo struct {
	Timestamp   int64    `json:"timestamp"`
	AccountID   SecureID `json:"accountId"`
	Currency    string   `json:"currency"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Balance     float64  `json:"balance"`
	TotalLocked float64  `json:"totalLocked"`
}

// AccountResponse holds args for accounting requests
type AccountResponse struct {
	Accounts []AccountInfo `json:"accounts"`
}

// AccountingService operation return user's accounts
func (p *AccountingService) List(r *http.Request, request *AccountRequest, reply *AccountResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "AccountingService.List")
	log = GetServiceRequestLog(log, r, "Accounting", "List")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = getSessionCookie(r)
	sessionID := sessions.SessionID(request.SessionID)
	userID := session.UserSession(ctx, sessionID)
	if !sessions.IsUserValid(userID) {
		log.Error("Invalid userSession")
		return sessions.ErrInvalidSessionID
	}
	log = log.WithFields(logrus.Fields{
		"SessionID": sessionID,
		"UserID":    userID,
	})

	// call internal API
	list, err := client.AccountList(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("AccountList failed")
		return sessions.ErrInternalError
	}

	if list.UserID != userID {
		log.WithField("UserID", userID).
			Error("Wrong UserID")
		return sessions.ErrInternalError
	}

	// prepare response
	var result []AccountInfo
	for _, account := range list.Accounts {
		// create SecureID from AccountID
		result = append(result, AccountInfo{
			Timestamp:   makeTimestampMillis(account.Timestamp),
			AccountID:   secureAccountID(account.AccountID),
			Currency:    account.Currency,
			Name:        account.Name,
			Status:      account.Status,
			Balance:     account.Balance,
			TotalLocked: account.TotalLocked,
		})
	}

	// Reply
	*reply = AccountResponse{
		Accounts: result[:],
	}

	log.WithFields(logrus.Fields{
		"Count": len(reply.Accounts),
	}).Info("ListAccounts")

	return nil
}

// Todo use SecureID package
type SecureID string

func secureAccountID(accountID uint64) SecureID {
	return SecureID(utils.HashString(fmt.Sprintf("a:%d", accountID)))
}
