// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/secureid"

	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"
)

const (
	DefaulDepositCountByPage  = 50
	DefaulBatchCountByPage    = 50
	DefaulWithdrawCountByPage = 50
)

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

type TransferStatus struct {
	Deposit  DepositStatus  `json:"deposit"`
	Batch    BatchStatus    `json:"batch"`
	Withdraw WithdrawStatus `json:"withdraw"`
}

func FetchTransferStatus(ctx context.Context) (TransferStatus, error) {
	db := appcontext.Database(ctx)

	batchs, err := database.BatchsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	deposits, err := database.DepositsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	witdthdraws, err := database.WithdrawsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	return TransferStatus{
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
	}, nil
}

// DepositListRequest holds args for depositlist requests
type DepositListRequest struct {
	apiservice.SessionArgs
	RequestPaging
}

type DepositInfo struct {
	DepositID string  `json:"depositId"`
	Timestamp int64   `json:"timestamp"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}

// DepositListResponse holds response for depositlist request
type DepositListResponse struct {
	RequestPaging
	Deposits []DepositInfo `json:"deposits"`
}

func (p *DashboardService) DepositList(r *http.Request, request *DepositListRequest, reply *DepositListResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.DepositList")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "DepositList")

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

	sID := appcontext.SecureID(ctx)

	var startID secureid.Value
	if len(request.Start) > 0 {
		startID, err = sID.FromSecureID("deposit", sID.Parse(request.Start))
		if err != nil {
			log.WithError(err).
				WithField("Start", request.Start).
				Error("startID FromSecureID failed")
			return ErrPermissionDenied
		}
	}
	var pagesCount int
	var ids []model.OperationInfoID
	infos := make(map[model.OperationInfoID]DepositInfo)
	err = db.Transaction(func(db bank.Database) error {
		var err error
		pagesCount, err = database.DepositPagingCount(db, DefaulDepositCountByPage)
		if err != nil {
			pagesCount = 0
			return err
		}

		ids, err = database.DepositPage(db, model.OperationInfoID(startID), DefaulDepositCountByPage)
		if err != nil {
			ids = nil
			return err
		}

		for _, id := range ids {
			var info DepositInfo
			op, err := database.GetOperationInfo(db, id)
			if err != nil {
				ids = nil
				return err
			}
			status, err := database.GetOperationStatus(db, id)
			if err != nil {
				ids = nil
				return err
			}

			info.Timestamp = makeTimestampMillis(op.Timestamp)
			info.Status = status.Accounted
			info.Amount = float64(op.Amount)
			info.Currency = func() string {

				if op.AssetID == 0 {
					addr, err := database.GetCryptoAddress(db, op.CryptoAddressID)
					if err != nil {
						return ""
					}
					return getChainMainCurrency(ctx, string(addr.Chain))

				} else {
					asset, err := database.GetAsset(db, op.AssetID)
					if err != nil {
						return ""
					}
					return string(asset.CurrencyName)
				}
			}()

			infos[id] = info
		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("DepositPage failed")
		return apiservice.ErrServiceInternalError
	}

	var next string
	if len(ids) > 0 {
		nextID := int(ids[len(ids)-1]) + 1
		secureID, err := sID.ToSecureID("deposit", secureid.Value(uint64(nextID)))
		if err != nil {
			return err
		}
		next = sID.ToString(secureID)
	}

	var deposits []DepositInfo
	for _, id := range ids {
		secureID, err := sID.ToSecureID("deposit", secureid.Value(uint64(id)))
		if err != nil {
			return err
		}

		var depositInfo DepositInfo
		if info, ok := infos[id]; ok {
			depositInfo = info
		}
		depositInfo.DepositID = sID.ToString(secureID)

		deposits = append(deposits, depositInfo)
	}

	*reply = DepositListResponse{
		RequestPaging: RequestPaging{
			Page:         request.Page,
			PageCount:    pagesCount,
			CountPerPage: DefaulDepositCountByPage,
			Start:        request.Start,
			Next:         next,
		},
		Deposits: deposits,
	}

	return nil
}

// BatchListRequest holds args for batchlist requests
type BatchListRequest struct {
	apiservice.SessionArgs
	RequestPaging
}

type BatchInfo struct {
	BatchID      string `json:"batchId"`
	Timestamp    int64  `json:"timestamp"`
	ExecuteAfter int64  `json:"executeAfter"`
	Withdraws    int    `json:"withdraws"`
	Status       string `json:"status"`
}

// BatchListResponse holds response for batchlist request
type BatchListResponse struct {
	RequestPaging
	Batchs []BatchInfo `json:"batchs"`
}

func (p *DashboardService) BatchList(r *http.Request, request *BatchListRequest, reply *BatchListResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.BatchList")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "BatchList")

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

	sID := appcontext.SecureID(ctx)

	var startID secureid.Value
	if len(request.Start) > 0 {
		startID, err = sID.FromSecureID("batch", sID.Parse(request.Start))
		if err != nil {
			log.WithError(err).
				WithField("Start", request.Start).
				Error("startID FromSecureID failed")
			return ErrPermissionDenied
		}
	}
	var pagesCount int
	var ids []model.BatchID
	infos := make(map[model.BatchID]BatchInfo)
	err = db.Transaction(func(db bank.Database) error {
		var err error
		pagesCount, err = database.BatchPagingCount(db, DefaulBatchCountByPage)
		if err != nil {
			pagesCount = 0
			return err
		}

		ids, err = database.BatchPage(db, model.BatchID(startID), DefaulBatchCountByPage)
		if err != nil {
			ids = nil
			return err
		}
		for _, id := range ids {
			var info BatchInfo

			batch, err := database.GetBatch(db, id)
			if err != nil {
				ids = nil
				return err
			}
			status, err := database.GetLastBatchInfo(db, id)
			if err != nil {
				ids = nil
				return err
			}

			wids, err := database.GetBatchWithdraws(db, id)
			if err != nil {
				ids = nil
				return err
			}

			info.Timestamp = makeTimestampMillis(batch.Timestamp)
			info.ExecuteAfter = makeTimestampMillis(batch.ExecuteAfter)
			info.Withdraws = len(wids)
			info.Status = string(status.Status)

			infos[id] = info
		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("BatchPage failed")
		return apiservice.ErrServiceInternalError
	}

	var next string
	if len(ids) > 0 {
		nextID := int(ids[len(ids)-1]) + 1
		secureID, err := sID.ToSecureID("batch", secureid.Value(uint64(nextID)))
		if err != nil {
			return err
		}
		next = sID.ToString(secureID)
	}

	var batchs []BatchInfo
	for _, id := range ids {
		secureID, err := sID.ToSecureID("batch", secureid.Value(uint64(id)))
		if err != nil {
			return err
		}

		var batchInfo BatchInfo
		if info, ok := infos[id]; ok {
			batchInfo = info
		}
		batchInfo.BatchID = sID.ToString(secureID)

		batchs = append(batchs, batchInfo)
	}

	*reply = BatchListResponse{
		RequestPaging: RequestPaging{
			Page:         request.Page,
			PageCount:    pagesCount,
			CountPerPage: DefaulBatchCountByPage,
			Start:        request.Start,
			Next:         next,
		},
		Batchs: batchs,
	}

	return nil
}

// WithdrawListRequest holds args for withdrawlist requests
type WithdrawListRequest struct {
	apiservice.SessionArgs
	RequestPaging
}

type WithdrawInfo struct {
	WithdrawID string `json:"withdrawId"`
}

// BatchListResponse holds response for withdrawlist request
type WithdrawListResponse struct {
	RequestPaging
	Withdraws []WithdrawInfo `json:"withdraws"`
}

func (p *DashboardService) WithdrawList(r *http.Request, request *WithdrawListRequest, reply *WithdrawListResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.WithdrawList")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "WithdrawList")

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

	sID := appcontext.SecureID(ctx)

	var startID secureid.Value
	if len(request.Start) > 0 {
		startID, err = sID.FromSecureID("withdraw", sID.Parse(request.Start))
		if err != nil {
			log.WithError(err).
				WithField("Start", request.Start).
				Error("startID FromSecureID failed")
			return ErrPermissionDenied
		}
	}
	var pagesCount int
	var ids []model.WithdrawID
	err = db.Transaction(func(db bank.Database) error {
		var err error
		pagesCount, err = database.WithdrawPagingCount(db, DefaulWithdrawCountByPage)
		if err != nil {
			pagesCount = 0
			return err
		}

		ids, err = database.WithdrawPage(db, model.WithdrawID(startID), DefaulWithdrawCountByPage)
		if err != nil {
			ids = nil
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("WithdrawPage failed")
		return apiservice.ErrServiceInternalError
	}

	var next string
	if len(ids) > 0 {
		nextID := int(ids[len(ids)-1]) + 1
		secureID, err := sID.ToSecureID("withdraw", secureid.Value(uint64(nextID)))
		if err != nil {
			return err
		}
		next = sID.ToString(secureID)
	}

	var withdraws []WithdrawInfo
	for _, id := range ids {
		secureID, err := sID.ToSecureID("withdraw", secureid.Value(uint64(id)))
		if err != nil {
			return err
		}

		withdraws = append(withdraws, WithdrawInfo{
			WithdrawID: sID.ToString(secureID),
		})
	}

	*reply = WithdrawListResponse{
		RequestPaging: RequestPaging{
			Page:         request.Page,
			PageCount:    pagesCount,
			CountPerPage: DefaulWithdrawCountByPage,
			Start:        request.Start,
			Next:         next,
		},
		Withdraws: withdraws,
	}

	return nil
}

func getChainMainCurrency(ctx context.Context, chain string) string {
	switch chain {
	case "bitcoin-mainnet":
		return "BTC"

	case "bitcoin-testnet":
		return "TBTC"

	case "liquid-mainnet":
		return "LBTC"

	default:
		return ""
	}
}
