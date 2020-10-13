// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/secureid"

	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/networking/sessions"

	accounting "github.com/condensat/bank-core/accounting/client"
	"github.com/condensat/bank-core/wallet/client"

	"github.com/sirupsen/logrus"
)

var (
	ErrWalletChainNotFoundError = errors.New("Chain Not Found")
	ErrInvalidPublicAddress     = errors.New("Invalid Public Address")
	ErrInvalidWithdraw          = errors.New("Invalid Withdraw")
)

type WalletService int

// WalletNextDepositRequest holds args for accounting requests
type WalletNextDepositRequest struct {
	sessions.SessionArgs
	AccountID string `json:"accountId"`
}

// WalletNextDepositResponse holds args for accounting requests
type WalletNextDepositResponse struct {
	Currency        string `json:"currency"`
	DisplayCurrency string `json:"displayCurrency"`
	PublicAddress   string `json:"publicAddress"`
	URL             string `json:"url"`
}

// WalletService operation return deposit address for account
func (p *WalletService) NextDeposit(r *http.Request, request *WalletNextDepositRequest, reply *WalletNextDepositResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "WalletService.NextDeposit")
	log = networking.GetServiceRequestLog(log, r, "Wallet", "NextDeposit")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = sessions.GetSessionCookie(r)
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

	sID := appcontext.SecureID(ctx)
	accountID, err := sID.FromSecureID("account", sID.Parse(request.AccountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("Wrong AccountID")
		return sessions.ErrInternalError
	}

	account, err := accounting.AccountInfo(ctx, uint64(accountID))
	if err != nil {
		log.WithError(err).Error("AccountInfo failed")
		return err
	}
	if !account.Currency.Crypto {
		log.WithField("AccountID", request.AccountID).
			Error("Non Crypto Account")
		return sessions.ErrInternalError
	}
	chain, err := getChainFromCurrencyName(account.Currency.Crypto, account.Currency.Name)
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("getChainFromCurrencyName failed")
		return sessions.ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"Chain":     chain,
		"AccountID": accountID,
	})

	addr, err := client.CryptoAddressNextDeposit(ctx, chain, uint64(accountID))
	if err != nil {
		log.WithError(err).
			Error("CryptoAddressNextDeposit Failed")
		return ErrServiceInternalError
	}

	// Reply
	protocol, err := getProtocolFromCurrencyName(account.Currency.Crypto, account.Currency.Name)
	*reply = WalletNextDepositResponse{
		Currency:        account.Currency.Name,
		DisplayCurrency: account.Currency.DisplayName,
		PublicAddress:   addr.PublicAddress,
		URL:             fmt.Sprintf("%s:%s", protocol, addr.PublicAddress),
	}
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			WithField("CurrencyName", account.Currency.Name).
			Error("getProtocolFromCurrencyName failed")
		return sessions.ErrInternalError
	}

	log.WithFields(logrus.Fields{
		"CurrencyName":  reply.Currency,
		"PublicAddress": reply.PublicAddress,
		"Url":           reply.URL,
	}).Info("CryptoAddressNextDeposit")

	return nil
}

func getChainFromCurrencyName(isCrypto bool, currencyName string) (string, error) {
	switch currencyName {
	case "BTC":
		return "bitcoin-mainnet", nil
	case "TBTC":
		return "bitcoin-testnet", nil
	case "LBTC":
		return "liquid-mainnet", nil

	default:
		if isCrypto {
			return "liquid-mainnet", nil
		}
		return "", ErrWalletChainNotFoundError
	}
}

// WalletSendFundsRequest holds args for wallet requests
type WalletSendFundsRequest struct {
	sessions.SessionArgs
	AccountID     string  `json:"accountId"`
	PublicAddress string  `json:"publicAddress"`
	Amount        float64 `json:"amount"`
}

// WalletSendFundsResponse holds args for wallet requests
type WalletSendFundsResponse struct {
	WithdrawID string `json:"withdrawId"`
}

func (p *WalletService) SendFunds(r *http.Request, request *WalletSendFundsRequest, reply *WalletSendFundsResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "WalletService.SendFunds")
	log = networking.GetServiceRequestLog(log, r, "Wallet", "SendFunds")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = sessions.GetSessionCookie(r)
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

	sID := appcontext.SecureID(ctx)
	accountID, err := sID.FromSecureID("account", sID.Parse(request.AccountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("Wrong AccountID")
		return sessions.ErrInternalError
	}

	account, err := accounting.AccountInfo(ctx, uint64(accountID))
	if err != nil {
		log.WithError(err).Error("AccountInfo failed")
		return err
	}
	if account.Status != "normal" {
		log.WithFields(logrus.Fields{
			"AccountID": request.AccountID,
			"Status":    account.Status,
		}).Error("Account status does not allow to send fund")
		return ErrInvalidAccountID
	}
	if !account.Currency.Crypto {
		log.WithField("AccountID", request.AccountID).
			Error("Non Crypto Account")
		return sessions.ErrInternalError
	}
	chain, err := getChainFromCurrencyName(account.Currency.Crypto, account.Currency.Name)
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("getChainFromCurrencyName failed")
		return sessions.ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"Chain":     chain,
		"AccountID": accountID,
	})

	addr, err := client.AddressInfo(ctx, chain, request.PublicAddress)
	if err != nil {
		log.WithError(err).
			Error("AddressInfo Failed")
		return ErrInvalidPublicAddress
	}

	if !addr.IsValid {
		log.WithError(ErrInvalidPublicAddress).
			Error("PublicAddress is not valid")
		return ErrInvalidPublicAddress
	}

	log.WithFields(logrus.Fields{
		"Account": account,
		"Address": addr,
	}).Debug("Account and address infos")

	withdrawID, err := accounting.AccountTransferWithdrawCrypto(ctx, account.AccountID, account.Currency.DatabaseName, request.Amount, "normal", "Api SendFunds", chain, request.PublicAddress)
	if err != nil {
		log.WithError(err).
			Error("AccountTransferWithdrawCrypto Failed")
		return ErrInvalidWithdraw
	}

	secureID, err := sID.ToSecureID("withdraw", secureid.Value(withdrawID))
	if err != nil {
		log.WithError(err).
			Error("ToSecureID Failed")
		return ErrServiceInternalError
	}

	*reply = WalletSendFundsResponse{
		WithdrawID: sID.ToString(secureID),
	}

	return nil
}

// WalletCancelWithdrawRequest holds args for wallet requests
type WalletCancelWithdrawRequest struct {
	sessions.SessionArgs
	WithdrawID string `json:"withdrawId"`
}

// WalletCancelWithdrawResponse holds args for wallet requests
type WalletCancelWithdrawResponse struct {
	WithdrawID string `json:"withdrawId"`
	Status     string `json:"status"`
}

func (p *WalletService) CancelWithdraw(r *http.Request, request *WalletCancelWithdrawRequest, reply *WalletCancelWithdrawResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "WalletService.CancelWithdraw")
	log = networking.GetServiceRequestLog(log, r, "Wallet", "CancelWithdraw")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = sessions.GetSessionCookie(r)
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

	sID := appcontext.SecureID(ctx)
	withdrawID, err := sID.FromSecureID("withdraw", sID.Parse(request.WithdrawID))
	if err != nil {
		log.WithError(err).
			WithField("WithdrawID", request.WithdrawID).
			Error("Wrong WithdrawID")
		return sessions.ErrInternalError
	}

	log = log.WithField("WithdrawID", withdrawID)

	wi, err := accounting.CancelWithdraw(ctx, uint64(withdrawID))
	if err != nil {
		log.WithError(err).Error("CancelWithdraw failed")
		return err
	}

	*reply = WalletCancelWithdrawResponse{
		WithdrawID: request.WithdrawID,
		Status:     wi.Status,
	}

	return nil
}

// WalletSendHistoryRequest holds args for wallet requests
type WalletSendHistoryRequest struct {
	sessions.SessionArgs
}

type WithdrawInfo struct {
	WithdrawID string  `json:"withdrawId"`
	Timestamp  int64   `json:"timestamp"`
	AccountID  string  `json:"accountId"`
	Amount     float64 `json:"amount"`
	Chain      string  `json:"chain"`
	PublicKey  string  `json:"publicKey"`
	Status     string  `json:"status"`
}

// WalletSendFundsResponse holds args for wallet requests
type WalletSendHistoryResponse struct {
	Withdraws []WithdrawInfo `json:"withdraws"`
}

func (p *WalletService) SendHistory(r *http.Request, request *WalletSendHistoryRequest, reply *WalletSendHistoryResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "WalletService.SendHistory")
	log = networking.GetServiceRequestLog(log, r, "Wallet", "SendHistory")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = sessions.GetSessionCookie(r)
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

	userWithdraws, err := accounting.UserWithdrawsCrypto(ctx, userID)
	if !sessions.IsUserValid(userID) {
		log.WithError(err).
			Error("UserWithdrawsCrypto Failed")
		return sessions.ErrInvalidUserID
	}

	sID := appcontext.SecureID(ctx)

	var withdraws []WithdrawInfo
	for _, uw := range userWithdraws.Withdraws {
		swID, err := sID.ToSecureID("withdraw", secureid.Value(uw.WithdrawID))
		if err != nil {
			log.WithError(err).
				WithField("WithdrawID", uw.WithdrawID).
				Error("ToSecureID Failed")
			continue
		}
		saID, err := sID.ToSecureID("account", secureid.Value(uw.AccountID))
		if err != nil {
			log.WithError(err).
				WithField("AccountID", uw.AccountID).
				Error("ToSecureID Failed")
			continue
		}

		withdraws = append(withdraws, WithdrawInfo{
			WithdrawID: sID.ToString(swID),
			Timestamp:  makeTimestampMillis(uw.Timestamp),
			AccountID:  sID.ToString(saID),
			Amount:     uw.Amount,
			Chain:      uw.Chain,
			PublicKey:  uw.PublicKey,
			Status:     uw.Status,
		})
	}

	*reply = WalletSendHistoryResponse{
		Withdraws: withdraws[:],
	}

	return nil
}

func getProtocolFromCurrencyName(isCrypto bool, currencyName string) (string, error) {
	switch currencyName {
	case "BTC":
		return "bitcoin", nil
	case "TBTC":
		return "bitcoin-testnet", nil
	case "LBTC":
		return "liquid", nil

	default:
		if isCrypto {
			return "liquid", nil
		}
		return "", ErrWalletChainNotFoundError
	}
}
