// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"net/http"

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

	sID := appcontext.SecureID(ctx)

	// prepare response
	var result []AccountInfo
	for _, account := range list.Accounts {
		// create SecureID from AccountID
		secureID, err := sID.ToSecureID("account", secureid.Value(account.AccountID))
		if err != nil {
			log.WithError(err).
				WithField("AccountID", account.AccountID).
				Error("ToSecureID failed")
			return sessions.ErrInternalError
		}

		result = append(result, AccountInfo{
			Timestamp:   makeTimestampMillis(account.Timestamp),
			AccountID:   sID.ToString(secureID),
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

	sID := appcontext.SecureID(ctx)
	accountID, err := sID.FromSecureID("account", sID.Parse(request.AccountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("Wrong AccountID")
		return sessions.ErrInternalError
	}

	// call internal API
	from := fromTimestampMillis(request.From)
	to := fromTimestampMillis(request.To)

	history, err := client.AccountHistory(ctx, uint64(accountID), from, to)
	if err != nil {
		log.WithError(err).
			Error("AccountHistory failed")
		return sessions.ErrInternalError
	}

	if history.AccountID != uint64(accountID) {
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
		secureID, err := sID.ToSecureID("operation", secureid.Value(entry.OperationID))
		if err != nil {
			log.WithError(err).
				WithField("OperationID", entry.OperationID).
				Error("ToSecureID failed")
			return sessions.ErrInternalError
		}

		result = append(result, AccountOperation{
			Timestamp:   makeTimestampMillis(entry.Timestamp),
			OperationID: sID.ToString(secureID),
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
