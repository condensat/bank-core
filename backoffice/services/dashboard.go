// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"errors"
	"net/http"
	"sort"

	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/utils"

	wallet "github.com/condensat/bank-core/wallet/client"

	"github.com/condensat/bank-core/logger"
	logmodel "github.com/condensat/bank-core/logger/model"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

var (
	ErrPermissionDenied = errors.New("Permission Denied")
)

type DashboardService int

// StatusRequest holds args for status requests
type StatusRequest struct {
	apiservice.SessionArgs
}

type LogStatus struct {
	Warnings int `json:"warning"`
	Errors   int `json:"errors"`
	Panics   int `json:"panics"`
}

type UsersStatus struct {
	Count     int `json:"count"`
	Connected int `json:"connected"`
}

type CurrencyBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Locked   float64 `json:"locked,omitempty"`
}

type AccountingStatus struct {
	Count    int               `json:"count"`
	Active   int               `json:"active"`
	Balances []CurrencyBalance `json:"balances"`
}

type DepositStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type BatchStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type WithdrawStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type SwapStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type WalletInfo struct {
	UTXOs  int     `json:"utxos"`
	Amount float64 `json:"amount"`
}

type WalletStatus struct {
	Chain  string     `json:"chain"`
	Asset  string     `json:"asset"`
	Total  WalletInfo `json:"total"`
	Locked WalletInfo `json:"locked"`
}

type ReserveStatus struct {
	Wallets []WalletStatus `json:"wallets"`
}

// StatusResponse holds args for string requests
type StatusResponse struct {
	Logs       LogStatus        `json:"logs"`
	Users      UsersStatus      `json:"users"`
	Accounting AccountingStatus `json:"accounting"`
	Deposit    DepositStatus    `json:"deposit"`
	Batch      BatchStatus      `json:"batch"`
	Withdraw   WithdrawStatus   `json:"withdraw"`
	Swap       SwapStatus       `json:"swap"`
	Reserve    ReserveStatus    `json:"reserve"`
}

func (p *DashboardService) Status(r *http.Request, request *StatusRequest, reply *StatusResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.Status")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "Status")

	db := appcontext.Database(ctx)
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = apiservice.GetSessionCookie(r)
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

	isAdmin, err := database.UserHasRole(db, model.UserID(userID), model.RoleNameAdmin)
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

	logsInfo, err := logmodel.LogsInfo(db.DB().(*gorm.DB))
	if err != nil {
		log.WithError(err).
			Error("LogsInfo failed")
		return apiservice.ErrServiceInternalError
	}

	userCount, err := database.UserCount(db)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}
	sessionCount, err := session.Count(ctx)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}

	accountsInfo, err := database.AccountsInfos(db)
	if err != nil {
		log.WithError(err).
			Error("AccountInfos failed")
		return apiservice.ErrServiceInternalError
	}

	var balances []CurrencyBalance
	for _, account := range accountsInfo.Accounts {
		balances = append(balances, CurrencyBalance{
			Currency: account.CurrencyName,
			Balance:  account.Balance,
			Locked:   account.TotalLocked,
		})
	}

	batchs, err := database.BatchsInfos(db)
	if err != nil {
		log.WithError(err).
			Error("BatchsInfos failed")
		return apiservice.ErrServiceInternalError
	}

	deposits, err := database.DepositsInfos(db)
	if err != nil {
		log.WithError(err).
			Error("DepositsInfos failed")
		return apiservice.ErrServiceInternalError
	}

	witdthdraws, err := database.WithdrawsInfos(db)
	if err != nil {
		log.WithError(err).
			Error("WithdrawsInfos failed")
		return apiservice.ErrServiceInternalError
	}

	swaps, err := database.SwapssInfos(db)
	if err != nil {
		log.WithError(err).
			Error("SwapssInfos failed")
		return apiservice.ErrServiceInternalError
	}

	walletStatus, err := wallet.WalletStatus(ctx)
	if err != nil {
		log.WithError(err).
			Error("WalletStatus failed")
		return apiservice.ErrServiceInternalError
	}

	var wallets []WalletStatus
	assetMap := make(map[string]*WalletStatus)
	for _, wallet := range walletStatus.Wallets {
		for _, utxo := range wallet.UTXOs {

			// get or create WalletStatus from assetMap
			key := wallet.Chain + utxo.Asset
			ws, ok := assetMap[key]
			if !ok {
				ws = &WalletStatus{
					Chain: wallet.Chain,
					Asset: utxo.Asset,
				}
				assetMap[key] = ws
			}

			ws.Total.Amount += utxo.Amount
			ws.Total.UTXOs++
			if utxo.Locked {
				ws.Locked.Amount += utxo.Amount
				ws.Locked.UTXOs++
			}
		}
	}

	for _, ws := range assetMap {
		ws.Total.Amount = utils.ToFixed(ws.Total.Amount, 8)
		ws.Locked.Amount = utils.ToFixed(ws.Locked.Amount, 8)

		wallets = append(wallets, *ws)
	}

	sort.Slice(wallets, func(i, j int) bool {
		if wallets[i].Chain != wallets[j].Chain {
			return wallets[i].Chain < wallets[j].Chain
		}

		return wallets[i].Asset < wallets[j].Asset
	})

	*reply = StatusResponse{
		Logs: LogStatus{
			Warnings: logsInfo.Warnings,
			Errors:   logsInfo.Errors,
			Panics:   logsInfo.Panics,
		},
		Users: UsersStatus{
			Count:     userCount,
			Connected: sessionCount,
		},
		Accounting: AccountingStatus{
			Count:    accountsInfo.Count,
			Active:   accountsInfo.Active,
			Balances: balances,
		},
		Deposit: DepositStatus{
			Count:      deposits.Count,
			Processing: deposits.Active,
		},
		Batch: BatchStatus{
			Count:      batchs.Count,
			Processing: batchs.Active,
		},
		Withdraw: WithdrawStatus{
			Count:      witdthdraws.Count,
			Processing: witdthdraws.Active,
		},
		Swap: SwapStatus{
			Count:      swaps.Count,
			Processing: swaps.Active,
		},
		Reserve: ReserveStatus{
			Wallets: wallets,
		},
	}

	log.WithFields(logrus.Fields{
		"UserCount":    userCount,
		"SessionCount": sessionCount,
	}).Info("Status")

	return nil
}
