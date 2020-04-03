// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/secureid"

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
	Timestamp   int64   `json:"timestamp"`
	AccountID   string  `json:"accountId"`
	Currency    string  `json:"currency"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Balance     float64 `json:"balance"`
	TotalLocked float64 `json:"totalLocked"`
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
			AccountID:   getSecureIDString(ctx, "account", account.AccountID),
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

// AccountHistoryRequest holds args for accounting history requests
type AccountHistoryRequest struct {
	SessionArgs
	AccountID string `json:"accountId"`
	From      int64  `json:"from"`
	To        int64  `json:"to"`
}

// AccountOperation holds account operation
type AccountOperation struct {
	Timestamp   int64   `json:"timestamp"`
	OperationID string  `json:"operationId"`
	Amount      float64 `json:"amount"`
	Balance     float64 `json:"balance"`
	LockAmount  float64 `json:"lockAmount"`
	TotalLocked float64 `json:"totalLocked"`
}

// AccountHistoryResponse holds args for accounting requests
type AccountHistoryResponse struct {
	AccountID  string             `json:"accountId"`
	Currency   string             `json:"currency"`
	From       int64              `json:"from"`
	To         int64              `json:"to"`
	Operations []AccountOperation `json:"operations"`
}

// AccountingService operation return user's accounts
func (p *AccountingService) History(r *http.Request, request *AccountHistoryRequest, reply *AccountHistoryResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "AccountingService.History")
	log = GetServiceRequestLog(log, r, "Accounting", "History")

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

	accountID := getIDFromSecureIDString(ctx, "account", request.AccountID)

	// call internal API
	from := fromTimestampMillis(request.From)
	to := fromTimestampMillis(request.To)

	history, err := client.AccountHistory(ctx, accountID, from, to)
	if err != nil {
		log.WithError(err).
			Error("AccountHistory failed")
		return sessions.ErrInternalError
	}

	if history.AccountID != accountID {
		log.WithField("AccountID", accountID).
			Error("Wrong AccountID")
		return sessions.ErrInternalError
	}

	// prepare response
	var result []AccountOperation
	// initialize date range with first entry
	if len(history.Entries) > 0 {
		from = history.Entries[0].Timestamp
		to = history.Entries[0].Timestamp
	}
	for _, entry := range history.Entries {
		// update date range from entry timestamp
		if from.After(entry.Timestamp) {
			from = entry.Timestamp
		}
		if to.Before(entry.Timestamp) {
			to = entry.Timestamp
		}
		// create SecureID from OperationID
		result = append(result, AccountOperation{
			Timestamp:   makeTimestampMillis(entry.Timestamp),
			OperationID: getSecureIDString(ctx, "operation", entry.OperationID),
			Amount:      entry.Amount,
			Balance:     entry.Balance,
			LockAmount:  entry.LockAmount,
			TotalLocked: entry.TotalLocked,
		})
	}

	// Reply
	*reply = AccountHistoryResponse{
		AccountID: request.AccountID,
		Currency:  history.Currency,
		From:      makeTimestampMillis(from),
		To:        makeTimestampMillis(to),

		Operations: result[:],
	}

	log.WithFields(logrus.Fields{
		"From":  reply.From,
		"To":    reply.To,
		"Count": len(reply.Operations),
	}).Info("Account History")

	return nil
}

func getSecureIDString(ctx context.Context, prefix string, value uint64) string {
	log := logger.Logger(ctx).WithField("Method", "getSecureIDString")
	sID := appcontext.SecureID(ctx)

	secureID, err := sID.ToSecureID(prefix, secureid.Value(value))
	if err != nil {
		log.WithError(err).
			WithField("Value", value).
			Error("ToSecureID failed")
		return ""
	}

	return fmt.Sprintf("%s:%s:%s", secureID.Version, secureID.Data, secureID.Check)
}

func getIDFromSecureIDString(ctx context.Context, prefix string, secureID string) uint64 {
	log := logger.Logger(ctx).WithField("Method", "getIDFromSecureIDString")
	sID := appcontext.SecureID(ctx)

	toks := strings.Split(secureID, ":")
	if len(toks) != 3 {
		return 0
	}

	value, err := sID.FromSecureID(prefix, secureid.SecureID{
		Version: toks[0],
		Data:    toks[1],
		Check:   toks[2],
	})
	if err != nil {
		log.WithError(err).
			WithField("SecureID", secureID).
			Error("FromSecureID failed")
		return 0
	}

	return uint64(value)
}
