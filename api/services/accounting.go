// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"math"
	"net/http"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/currency/rate"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"
	"github.com/condensat/secureid"

	"github.com/condensat/bank-core/accounting/client"
	"github.com/condensat/bank-core/api/sessions"

	"github.com/sirupsen/logrus"
)

type AccountingService int

// AccountRequest holds args for accounting requests
type AccountRequest struct {
	SessionArgs
	RateBase        string `json:"rateBase"`
	WithEmptyCrypto bool   `json:"withEmptyCrypto"`
}

type CurrencyInfo struct {
	DisplayName      string `json:"displayName"`
	Ticker           string `json:"ticker"`
	IsCrypto         bool   `json:"isCrypto"`
	IsAsset          bool   `json:"isAsset"`
	AssetHash        string `json:"assetHash"`
	DisplayPrecision uint   `json:"displayPrecision"`
	Icon             []byte `json:"icon,omitempty"`
}

type Notional struct {
	RateBase         string  `json:"rateBase"`
	DisplayPrecision uint    `json:"displayPrecision"`
	Rate             float64 `json:"rate"`
	Balance          float64 `json:"balance"`
	TotalLocked      float64 `json:"totalLocked"`
}

// AccountInfo holds account information
type AccountInfo struct {
	Timestamp   int64        `json:"timestamp"`
	AccountID   string       `json:"accountId"`
	Currency    CurrencyInfo `json:"curency"`
	Name        string       `json:"name"`
	Status      string       `json:"status"`
	Balance     float64      `json:"balance"`
	TotalLocked float64      `json:"totalLocked"`
	Notional    Notional     `json:"notional"`
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
	request.SessionID = GetSessionCookie(r)
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

		// compute convertion rate
		dstRate := 1.0
		// default RateBase to CHF
		if len(request.RateBase) == 0 {
			request.RateBase = "CHF"
		}
		if request.RateBase != "USD" {
			dst, err := rate.FetchRedisRate(ctx, request.RateBase, "USD")
			if err != nil {
				log.WithError(err).
					WithField("CurrencyName", request.RateBase).
					Error("FetchRedisRate failed")
				// non fatal, continue
				request.RateBase = ""
			}

			if dst.Rate > 0.0 {
				dstRate = 1.0 / dst.Rate
			}
		}

		finaleRate := 1.0
		if account.Currency.Name != "USD" {
			currencyRate, err := rate.FetchRedisRate(ctx, account.Currency.Name, "USD")
			if err != nil {
				log.WithError(err).
					WithField("CurrencyName", account.Currency.Name).
					Error("FetchRedisRate failed")
				// non fatal, continue
				currencyRate.Rate = 1.0
			}
			finaleRate = currencyRate.Rate * dstRate
		}

		info, err := rate.CurrencyInfo(ctx, request.RateBase)
		if err != nil {
			log.WithError(err).
				WithField("CurrencyName", account.Currency.Name).
				Error("FetchRedisRate failed")
			// non fatal, continue
			info.DisplayPrecision = account.Currency.DisplayPrecision
		}

		info.Asset = account.Currency.Type == 2 && account.Currency.Name != "LBTC"

		if info.Asset {
			finaleRate = 1.0
		}

		notional := Notional{
			RateBase:         request.RateBase,
			DisplayPrecision: info.DisplayPrecision,
			Rate:             utils.ToFixed(finaleRate, 12), // maximum precision for rates
			Balance:          utils.ToFixed(account.Balance/finaleRate, int(info.DisplayPrecision)),
			TotalLocked:      utils.ToFixed(account.TotalLocked/finaleRate, int(info.DisplayPrecision)),
		}

		if info.Asset || account.Currency.Name == "TBTC" {
			// asset and TBTC does not have notional
			notional = Notional{}
		}

		icon := getTickerIcon(ctx, account.Currency.Name)

		displayName := account.Currency.DisplayName
		if account.Currency.Name == "TBTC" {
			displayName = "Bitcoin testnet"
		}

		var assetHash string
		if info.Asset {
			if asset, err := database.GetAssetByCurrencyName(appcontext.Database(ctx), model.CurrencyName(account.Currency.Name)); err == nil {
				assetHash = string(asset.Hash)
			}
		}

		result = append(result, AccountInfo{
			Timestamp: makeTimestampMillis(account.Timestamp),
			AccountID: sID.ToString(secureID),
			Currency: CurrencyInfo{
				DisplayName:      displayName,
				Ticker:           account.Currency.Name,
				IsCrypto:         account.Currency.Crypto,
				IsAsset:          info.Asset,
				AssetHash:        assetHash,
				DisplayPrecision: account.Currency.DisplayPrecision,
				Icon:             icon,
			},
			Name:        account.Name,
			Status:      account.Status,
			Balance:     account.Balance,
			TotalLocked: account.TotalLocked,
			Notional:    notional,
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
	WithEmpty bool   `json:"withEmpty"`
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
	AccountID   string             `json:"accountId"`
	DisplayName string             `json:"displayName"`
	Ticker      string             `json:"ticker"`
	From        int64              `json:"from"`
	To          int64              `json:"to"`
	Operations  []AccountOperation `json:"operations"`
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
	request.SessionID = GetSessionCookie(r)
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
		// skip entries with zero amount (unlock operations)
		if !request.WithEmpty && math.Abs(entry.Amount) <= 0.0 {
			continue
		}
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

	displayName := history.DisplayName
	if history.DisplayName == "TBTC" {
		displayName = "Bitcoin testnet"
	}

	// Reply
	*reply = AccountHistoryResponse{
		AccountID:   request.AccountID,
		DisplayName: displayName,
		Ticker:      history.Ticker,
		From:        makeTimestampMillis(from),
		To:          makeTimestampMillis(to),

		Operations: result[:],
	}

	log.WithFields(logrus.Fields{
		"From":  reply.From,
		"To":    reply.To,
		"Count": len(reply.Operations),
	}).Info("Account History")

	return nil
}
