// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"net/http"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/networking"

	"github.com/condensat/bank-core/networking/sessions"

	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/database/model"
)

// WalletListRequest holds args for walletlist requests
type WalletListRequest struct {
	apiservice.SessionArgs
}

type WalletListStatus struct {
	Wallet string `json:"wallet"`
	UTXOs  int    `json:"utxos"`
	Height int    `json:"height"`
	Status string `json:"status"`
}

// BatchListResponse holds response for walletlist request
type WalletListResponse struct {
	Wallets []WalletListStatus `json:"wallets"`
}

func (p *DashboardService) WalletList(r *http.Request, request *WalletListRequest, reply *WalletListResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.WalletList")
	log = networking.GetServiceRequestLog(log, r, "Dashboard", "WalletList")

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

	wallets, err := FetchWalletList(ctx)
	if err != nil {
		log.WithError(err).
			Error("FetchWalletList failed")
		return sessions.ErrInternalError
	}

	var walletsStatus []WalletListStatus
	for _, wallet := range wallets {

		// request wallet detail
		var height int
		var utxos int
		walletDetail, err := FetchWalletDetail(ctx, wallet)
		if err != nil {
			log.WithError(err).
				Error("FetchWalletDetail failed")
			return sessions.ErrInternalError
		}
		if len(walletDetail) == 1 {
			height = walletDetail[0].Height
			utxos = len(walletDetail[0].UTXOs)
		}

		walletsStatus = append(walletsStatus, WalletListStatus{
			Wallet: wallet,
			UTXOs:  utxos,
			Height: height,
			Status: "sync",
		})
	}

	*reply = WalletListResponse{
		Wallets: walletsStatus,
	}

	return nil
}

// WalletDetailRequest holds args for walletdetail requests
type WalletDetailRequest struct {
	apiservice.SessionArgs
	Wallet string
}

// WalletDetailResponse holds response for walletdetail request
type WalletDetailResponse struct {
	Wallet string       `json:"wallet"`
	UTXOs  []WalletUTXO `json:"utxos"`
}

func (p *DashboardService) WalletDetail(r *http.Request, request *WalletDetailRequest, reply *WalletDetailResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.WalletDetail")
	log = networking.GetServiceRequestLog(log, r, "Dashboard", "WalletDetail")

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

	if len(request.Wallet) == 0 {
		if err != nil {
			log.
				Error("Invalid wallet")
			return sessions.ErrInternalError
		}
	}

	// request wallet detail
	wallets, err := FetchWalletDetail(ctx, request.Wallet)
	if err != nil {
		log.WithError(err).
			Error("FetchWalletList failed")
		return sessions.ErrInternalError
	}

	// only one wallet was requested
	if len(wallets) != 1 {
		log.
			Error("Invalid wallet detail")
		return sessions.ErrInternalError
	}

	walletInfo := wallets[0]

	*reply = WalletDetailResponse{
		Wallet: walletInfo.Chain,
		UTXOs:  walletInfo.UTXOs,
	}

	return nil
}
