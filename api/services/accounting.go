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
	"github.com/condensat/bank-core/accounting/common"
	"github.com/condensat/bank-core/api/sessions"

	"github.com/shengdoushi/base58"
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

// AccountHistoryRequest holds args for accounting history requests
type AccountHistoryRequest struct {
	SessionArgs
	AccountID SecureID `json:"accountId"`
	From      int64    `json:"from"`
	To        int64    `json:"to"`
}

// AccountOperation holds account operation
type AccountOperation struct {
	Timestamp   int64    `json:"timestamp"`
	OperationID SecureID `json:"operationId"`
	Amount      float64  `json:"amount"`
	Balance     float64  `json:"balance"`
	LockAmount  float64  `json:"lockAmount"`
	TotalLocked float64  `json:"totalLocked"`
}

// AccountHistoryResponse holds args for accounting requests
type AccountHistoryResponse struct {
	AccountID  SecureID           `json:"accountId"`
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

	// Todo use SecureID package
	accountID := accountIDFromSecureID(request.AccountID, list.Accounts)

	// call internal API
	history, err := client.AccountHistory(ctx, accountID, fromTimestampMillis(request.From), fromTimestampMillis(request.To))
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
	for _, entry := range history.Entries {
		// create SecureID from OperationID
		result = append(result, AccountOperation{
			Timestamp:   makeTimestampMillis(entry.Timestamp),
			OperationID: secureOperationID(entry.OperationID),
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
		From:      request.From,
		To:        request.To,

		Operations: result[:],
	}

	log.WithFields(logrus.Fields{
		"From":  reply.From,
		"To":    reply.To,
		"Count": len(reply.Operations),
	}).Info("Account History")

	return nil
}

// Todo use SecureID package
type SecureID string

// Todo use SecureID package
func secureIDString(prefix string, operationId uint64) SecureID {
	hash := utils.HashString(fmt.Sprintf("%s:%d", prefix, operationId))
	return SecureID(base58.Encode(hash, base58.BitcoinAlphabet))
}

// Todo use SecureID package
func secureAccountID(accountID uint64) SecureID {
	return secureIDString("a", accountID)
}

// Todo use SecureID package
func accountIDFromSecureID(sID SecureID, accounts []common.AccountInfo) uint64 {
	for _, account := range accounts {
		accountID := uint64(account.AccountID)
		if secureAccountID(accountID) == sID {
			return accountID
		}
	}
	return 0
}

// Todo use SecureID package
func secureOperationID(operationId uint64) SecureID {
	return secureIDString("o", operationId)
}
