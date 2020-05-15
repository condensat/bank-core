// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"errors"
	"net/http"

	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/secureid"

	accounting "github.com/condensat/bank-core/accounting/client"
	wallet "github.com/condensat/bank-core/wallet/client"

	"github.com/condensat/bank-core/swap/liquid/client"
	"github.com/condensat/bank-core/swap/liquid/common"

	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidAccountID = errors.New("Invalid AccountID")

	ErrInvalidProposal       = errors.New("Invalid Proposal")
	ErrInsufficientFunds     = errors.New("Insufficient funds")
	ErrInvalidProposerAsset  = errors.New("Invalid Proposer Asset")
	ErrInvalidProposerAmount = errors.New("Invalid Proposer Amount")
	ErrInvalidReceiverAsset  = errors.New("Invalid Receiver Asset")
	ErrInvalidReceiverAmount = errors.New("Invalid Receiver Amount")
)

type SwapService int

type ProposalInfo struct {
	ProposerAssetID string  `json:"proposerAssetId,omitempty"`
	ProposerAmount  float64 `json:"proposerAmount,omitempty"`
	ReceiverAssetID string  `json:"receiverAssetId,omitempty"`
	ReceiverAmount  float64 `json:"receiverAmount,omitempty"`
}

// SwapProposeRequest holds args for swap requests
type SwapRequest struct {
	SessionArgs
	AccountID string `json:"accountId"`
	SwapID    string `json:"swapId"`
	Payload   string `json:"payload,omitempty"`
}

// SwapProposeRequest holds args for swap requests
type SwapProposeRequest struct {
	SwapRequest
	Proposal ProposalInfo `json:"proposal,omitempty"`
}

// SwapProposeResponse holds args for swap reqponses
type SwapResponse struct {
	SwapID  string `json:"swapId"`
	Payload string `json:"payload,omitempty"`
}

// Propose operation return proposal payload for swap
func (p *SwapService) Propose(r *http.Request, request *SwapProposeRequest, reply *SwapResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "SwapService.Propose")
	log = GetServiceRequestLog(log, r, "swap", "Propose")

	proposal := request.Proposal

	if len(proposal.ProposerAssetID) == 0 {
		return ErrInvalidProposerAsset
	}
	if proposal.ProposerAmount <= 0.0 {
		return ErrInvalidProposerAmount
	}
	if len(proposal.ReceiverAssetID) == 0 {
		return ErrInvalidProposerAsset
	}
	if proposal.ReceiverAmount <= 0.0 {
		return ErrInvalidReceiverAmount
	}
	if proposal.ProposerAssetID == proposal.ReceiverAssetID {
		return ErrInvalidProposal
	}

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

	db := appcontext.Database(ctx)

	account, err := accounting.AccountInfo(ctx, uint64(accountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", accountID).
			Error("AccountInfo failed")
		return ErrInvalidAccountID
	}
	if !account.Currency.Crypto && !account.Currency.Asset {
		log.WithField("AccountID", request.AccountID).
			Error("Non Asset Account")
		return sessions.ErrInternalError
	}

	assetAccount, err := database.GetAssetByCurrencyName(db, model.CurrencyName(account.Currency.Name))
	if err != nil {
		log.WithField("CurrencyName", account.Currency.Name).
			Error("GetAssetByCurrencyName failed")
		return sessions.ErrInternalError
	}

	// ovveride ReceiverAssetID with user account asset
	if proposal.ReceiverAssetID != string(assetAccount.Hash) {
		return ErrInvalidReceiverAsset
	}

	// check for balances
	if (account.Balance - account.TotalLocked) < proposal.ProposerAmount {
		log.WithFields(logrus.Fields{
			"AccountID":      request.AccountID,
			"Balance":        account.Balance,
			"TotalLocked":    account.TotalLocked,
			"ProposerAmount": proposal.ProposerAmount,
		}).Error("Insufficient Funds")
		return ErrInsufficientFunds
	}

	chain, err := getChainFromCurrencyName(account.Currency.Crypto, account.Currency.Name)
	if err != nil {
		log.WithError(err).
			Error("getChainFromCurrencyName failed")
		return sessions.ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"Chain":        chain,
		"AccountID":    accountID,
		"CurrencyName": account.Currency.Name,
	})

	// generate new dedicated CryptoAddress
	addr, err := wallet.CryptoAddressNewDeposit(ctx, chain, uint64(accountID))
	if err != nil {
		log.WithError(err).
			Error("CryptoAddressNewDeposit Failed")
		return ErrServiceInternalError
	}

	amountDebit := model.Float(proposal.ProposerAmount)
	assetDebit, err := database.GetAssetByHash(db, model.AssetHash(proposal.ProposerAssetID))
	if err != nil {
		log.WithError(err).
			Error("GetAssetByHash Failed for Proposer")
		return ErrInvalidProposerAsset
	}
	amountCredit := model.Float(proposal.ReceiverAmount)
	assetCredit, err := database.GetAssetByHash(db, model.AssetHash(proposal.ReceiverAssetID))
	if err != nil {
		log.WithError(err).
			Error("GetAssetByHash Failed for Receiver")
		return ErrInvalidProposerAsset
	}

	swap, err := database.AddSwap(db, model.SwapTypeAsk,
		model.CryptoAddressID(addr.CryptoAddressID),
		assetDebit.ID, amountDebit,
		assetCredit.ID, amountCredit,
	)
	if err != nil {
		log.WithError(err).
			Error("AddSwap Failed")
		return ErrInvalidProposal
	}

	swapID := uint64(swap.ID)
	secureID, err := sID.ToSecureID("swap", secureid.Value(swapID))
	if err != nil {
		log.WithError(err).
			WithField("swapID", swapID).
			Error("ToSecureID failed")
		return sessions.ErrInternalError
	}

	address := common.ConfidentialAddress(addr.PublicAddress)

	log = log.WithFields(logrus.Fields{
		"SwapID":       sID.ToString(secureID),
		"Address":      address,
		"DebitAsset":   swap.DebitAsset,
		"DebitAmount":  swap.DebitAmount,
		"CredidAsset":  swap.CreditAsset,
		"CreditAmount": swap.CreditAmount,
	})

	swapProposal, err := client.CreateSwapProposal(ctx, swapID, address, common.ProposalInfo{
		ProposerAsset:  common.AssetID(proposal.ProposerAssetID),
		ProposerAmount: proposal.ProposerAmount,
		ReceiverAsset:  common.AssetID(proposal.ReceiverAssetID),
		ReceiverAmount: proposal.ReceiverAmount,
	}, common.DefaultFeeRate)
	if err != nil {
		log.WithError(err).
			Error("CreateSwapProposal failed")
		return sessions.ErrInternalError
	}

	swapInfo, err := database.AddSwapInfo(db, model.SwapID(swapID), model.SwapStatusProposed, model.Payload(swapProposal.Payload))
	if err != nil {
		log.WithError(err).
			Error("AddSwapInfo failed")
		return sessions.ErrInternalError
	}

	// Reply
	*reply = SwapResponse{
		SwapID:  sID.ToString(secureID),
		Payload: string(swapProposal.Payload),
	}

	log.
		WithField("SwapInfoID", swapInfo.ID).
		Info("Proposed")

	return nil
}
