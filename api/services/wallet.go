// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"net/http"

	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/client"

	"github.com/sirupsen/logrus"
)

type WalletService int

// WalletNextDepositRequest holds args for accounting requests
type WalletNextDepositRequest struct {
	SessionArgs
	AccountID string `json:"accountId"`
}

// WalletNextDepositResponse holds args for accounting requests
type WalletNextDepositResponse struct {
	PublicAddress string `json:"public_address"`
}

// WalletService operation return deposit address for account
func (p *WalletService) NextDeposit(r *http.Request, request *WalletNextDepositRequest, reply *WalletNextDepositResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "WalletService.NextDeposit")
	log = GetServiceRequestLog(log, r, "Wallet", "NextDeposit")

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

	sID := appcontext.SecureID(ctx)
	accountID, err := sID.FromSecureID("account", sID.Parse(request.AccountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("Wrong AccountID")
		return sessions.ErrInternalError
	}

	// Todo: find chain from accountID
	var chain string

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
	*reply = WalletNextDepositResponse{
		PublicAddress: addr.PublicAddress,
	}

	log.WithFields(logrus.Fields{
		"PublicAddress": len(reply.PublicAddress),
	}).Info("CryptoAddressNextDeposit")

	return nil
}
