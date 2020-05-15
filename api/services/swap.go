// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"encoding/base64"
	"encoding/json"
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

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidAccountID = errors.New("Invalid AccountID")
	ErrInvalidSwapID    = errors.New("Invalid SwapID")
	ErrInvalidPayload   = errors.New("Invalid Payload")

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

// Info operation return proposal payload for swap
func (p *SwapService) Info(r *http.Request, request *SwapRequest, reply *SwapResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "SwapService.Info")
	log = GetServiceRequestLog(log, r, "swap", "Info")

	if len(request.Payload) == 0 {
		return ErrInvalidPayload
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

	db := appcontext.Database(ctx)

	var swapData SwapData
	decoded, err := base64.StdEncoding.DecodeString(string(request.Payload))
	if err != nil {
		log.WithError(err).
			Error("Decode request Payload failed")
		return sessions.ErrInternalError
	}

	err = json.Unmarshal(decoded, &swapData)
	if err != nil {
		log.WithError(err).
			Error("Unmarshal SwapData failed")
		return sessions.ErrInternalError
	}

	var swapID string
	// try to get swapID from Unconfidential address
	if len(swapData.ProposerUnconfidentialAddress) != 0 {
		addr, err := database.GetCryptoAddressWithUnconfidential(db, model.String(swapData.ProposerUnconfidentialAddress))
		if err != nil && err != gorm.ErrRecordNotFound {
			log.WithError(err).
				Error("GetCryptoAddressWithUnconfidential failed")
			return sessions.ErrInternalError
		}

		if addr.ID > 0 {
			swap, err := database.GetSwapByCryptoAddressID(db, addr.ID)
			if err != nil {
				log.WithError(err).
					Error("GetSwapByCryptoAddressID failed")
				return sessions.ErrInternalError
			}

			sID := appcontext.SecureID(ctx)
			secureID, err := sID.ToSecureID("swap", secureid.Value(uint64(swap.ID)))
			if err != nil {
				log.WithError(err).
					WithField("swapID", swapID).
					Error("ToSecureID failed")
				return sessions.ErrInternalError
			}
			// store swapID for reply
			swapID = sID.ToString(secureID)
		}
	}

	swapInfo, err := client.InfoSwapProposal(ctx, uint64(0), common.Payload(request.Payload))
	if err != nil {
		log.WithError(err).
			Error("InfoSwapProposal failed")
		return sessions.ErrInternalError
	}

	// Reply
	*reply = SwapResponse{
		SwapID:  swapID,
		Payload: string(swapInfo.Payload),
	}

	log.Info("Info")

	return nil
}

// Finalize operation return proposal payload for swap
func (p *SwapService) Finalize(r *http.Request, request *SwapRequest, reply *SwapResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "SwapService.Finalize")
	log = GetServiceRequestLog(log, r, "swap", "Finalize")

	if len(request.SwapID) == 0 {
		return ErrInvalidSwapID
	}
	if len(request.Payload) == 0 {
		return ErrInvalidPayload
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
	swapID, err := sID.FromSecureID("swap", sID.Parse(request.SwapID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("Wrong AccountID")
		return sessions.ErrInternalError
	}

	swapInfo, err := client.InfoSwapProposal(ctx, 0, common.Payload(request.Payload))
	if err != nil {
		log.WithError(err).
			Error("InfoSwapProposal failed")
		return sessions.ErrInternalError
	}

	var decodedInfo SwapInfo
	err = json.Unmarshal([]byte(swapInfo.Payload), &decodedInfo)
	if err != nil {
		log.WithError(err).
			Error("SwapInfo decoding failed")
		return ErrInvalidProposal
	}
	if decodedInfo.Status != "accepted" {
		log.WithError(err).
			Error("SwapInfo status is not accepted")
		return ErrInvalidProposal
	}

	db := appcontext.Database(ctx)

	swap, err := database.GetSwap(db, model.SwapID(swapID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("Wrong SwapID")
		return ErrInvalidSwapID
	}

	// check if userID & accountID match swap CryptoAddressID
	addr, err := database.GetCryptoAddress(db, swap.CryptoAddressID)
	if err != nil {
		log.WithError(err).
			WithField("CryptoAddressID", swap.CryptoAddressID).
			Error("GetCryptoAddress failed")
		return ErrInvalidSwapID
	}

	accountID := addr.AccountID

	account, err := accounting.AccountInfo(ctx, uint64(accountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", accountID).
			Error("AccountInfo failed")
		return ErrInvalidSwapID
	}
	if !account.Currency.Crypto && !account.Currency.Asset {
		log.WithField("AccountID", request.AccountID).
			Error("Non Asset Account")
		return ErrInvalidSwapID
	}

	accountAsset, err := database.GetAssetByCurrencyName(db, model.CurrencyName(account.Currency.Name))
	if err != nil {
		log.WithField("CurrencyName", account.Currency.Name).
			Error("GetAssetByCurrencyName failed")
		return sessions.ErrInternalError
	}

	legCredit := decodedInfo.CreditLeg()
	if accountAsset.Hash != model.AssetHash(legCredit.Asset) {
		log.
			WithField("AccountAsset", accountAsset.Hash).
			WithField("CreditAsset", legCredit.Asset).
			Error("Wrong Receiver AssetHash")
		return ErrInvalidProposal
	}

	log = log.WithFields(logrus.Fields{
		"SwapID":    swapID,
		"AccountID": accountID,
	})

	finalized, err := client.FinalizeSwapProposal(ctx, uint64(swapID), common.Payload(request.Payload))
	if err != nil {
		log.WithError(err).
			Error("FinalizeSwapProposal failed")
		return sessions.ErrInternalError
	}

	sInfo, err := database.AddSwapInfo(db, model.SwapID(swapID), model.SwapStatusFinalized, model.Payload(finalized.Payload))
	if err != nil {
		log.WithError(err).
			Error("AddSwapInfo failed")
		return sessions.ErrInternalError
	}

	// Reply
	*reply = SwapResponse{
		SwapID:  request.SwapID,
		Payload: string(finalized.Payload),
	}

	log.
		WithField("SwapInfoID", sInfo.ID).
		Info("Finalized")

	return nil
}

// Accept operation return proposal payload for swap
func (p *SwapService) Accept(r *http.Request, request *SwapRequest, reply *SwapResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "SwapService.Accept")
	log = GetServiceRequestLog(log, r, "swap", "Accept")

	if len(request.Payload) == 0 {
		return ErrInvalidPayload
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

	swapInfo, err := client.InfoSwapProposal(ctx, 0, common.Payload(request.Payload))
	if err != nil {
		log.WithError(err).
			Error("InfoSwapProposal failed")
		return sessions.ErrInternalError
	}

	var decodedInfo SwapInfo
	err = json.Unmarshal([]byte(swapInfo.Payload), &decodedInfo)
	if err != nil {
		log.WithError(err).
			Error("SwapInfo decoding failed")
		return ErrInvalidProposal
	}
	if decodedInfo.Status != "proposed" {
		log.WithError(err).
			Error("SwapInfo status is not proposed")
		return ErrInvalidProposal
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

	legCredit := decodedInfo.CreditLeg()
	assetCredit, err := database.GetAssetByHash(db, model.AssetHash(legCredit.Asset))
	if err != nil {
		log.WithError(err).
			Error("GetAssetByHash Failed for asset credit")
		return ErrInvalidProposerAsset
	}
	legDebit := decodedInfo.DebitLeg()
	assetDebit, err := database.GetAssetByHash(db, model.AssetHash(legDebit.Asset))
	if err != nil {
		log.WithError(err).
			Error("GetAssetByHash Failed for asset debit")
		return ErrInvalidProposerAsset
	}

	accountAsset, err := database.GetAssetByCurrencyName(db, model.CurrencyName(account.Currency.Name))
	if err != nil {
		log.WithField("CurrencyName", account.Currency.Name).
			Error("GetAssetByCurrencyName failed")
		return sessions.ErrInternalError
	}
	if accountAsset.Hash != model.AssetHash(legCredit.Asset) {
		log.
			WithField("AccountAsset", accountAsset.Hash).
			WithField("CreditAsset", legCredit.Asset).
			Error("Wrong Credit AssetHash")
		return ErrInvalidProposal
	}

	// check for balances
	if (account.Balance - account.TotalLocked) < legDebit.Amount {
		log.WithFields(logrus.Fields{
			"AccountID":   request.AccountID,
			"Balance":     account.Balance,
			"TotalLocked": account.TotalLocked,
			"DebitAmount": legDebit.Amount,
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

	swap, err := database.AddSwap(db, model.SwapTypeBid,
		model.CryptoAddressID(addr.CryptoAddressID),
		assetDebit.ID, model.Float(legDebit.Amount),
		assetCredit.ID, model.Float(legCredit.Amount),
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

	accepted, err := client.AcceptSwapProposal(ctx, uint64(swapID), address, common.Payload(request.Payload), common.DefaultFeeRate)
	if err != nil {
		log.WithError(err).
			Error("AcceptSwapProposal failed")
		return sessions.ErrInternalError
	}

	sInfo, err := database.AddSwapInfo(db, model.SwapID(swapID), model.SwapStatusAccepted, model.Payload(accepted.Payload))
	if err != nil {
		log.WithError(err).
			Error("AddSwapInfo failed")
		return sessions.ErrInternalError
	}

	// Reply
	*reply = SwapResponse{
		SwapID:  sID.ToString(secureID),
		Payload: string(accepted.Payload),
	}

	log.
		WithField("SwapInfoID", sInfo.ID).
		Info("Accepted")

	return nil
}

type SwapData struct {
	ProposerUnconfidentialAddress string `json:"u_address_p"`
}

type SwapLeg struct {
	Incoming bool    `json:"incoming"`
	Funded   bool    `json:"funded"`
	Asset    string  `json:"asset"`
	Amount   float64 `json:"amount"`
	Fee      float64 `json:"fee"`
}

type SwapInfo struct {
	Status string    `json:"status"`
	Legs   []SwapLeg `json:"legs"`
}

func (p *SwapInfo) CreditLeg() SwapLeg {
	return legIncoming(p.Legs, true)
}

func (p *SwapInfo) DebitLeg() SwapLeg {
	return legIncoming(p.Legs, false)
}

func legIncoming(legs []SwapLeg, incomming bool) SwapLeg {
	if len(legs) != 2 {
		return SwapLeg{}
	}

	for _, leg := range legs {
		if leg.Incoming == incomming {
			return leg
		}
	}
	return SwapLeg{}
}
