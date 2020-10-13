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

	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/networking/sessions"
)

const (
	DefaulSwapCountByPage = 50
)

type SwapStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

func FetchSwapStatus(ctx context.Context) (SwapStatus, error) {
	db := appcontext.Database(ctx)

	swaps, err := database.SwapssInfos(db)
	if err != nil {
		return SwapStatus{}, err
	}

	return SwapStatus{
		Count:      swaps.Count,
		Processing: swaps.Active,
	}, nil
}

// SwapListRequest holds args for swaplist requests
type SwapListRequest struct {
	sessions.SessionArgs
	RequestPaging
}

type SwapInfo struct {
	SwapID    string `json:"swapId"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}

// SwapListResponse holds response for swaplist request
type SwapListResponse struct {
	RequestPaging
	Swaps []SwapInfo `json:"swaps"`
}

func (p *DashboardService) SwapList(r *http.Request, request *SwapListRequest, reply *SwapListResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.SwapList")
	log = networking.GetServiceRequestLog(log, r, "Dashboard", "SwapList")

	// Get userID from session
	request.SessionID = sessions.GetSessionCookie(r)
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
		startID, err = sID.FromSecureID("swap", sID.Parse(request.Start))
		if err != nil {
			log.WithError(err).
				WithField("Start", request.Start).
				Error("startID FromSecureID failed")
			return sessions.ErrInternalError
		}
	}
	var pagesCount int
	var ids []model.SwapID
	infos := make(map[model.SwapID]SwapInfo)
	err = db.Transaction(func(db bank.Database) error {
		var err error
		pagesCount, err = database.SwapPagingCount(db, DefaulSwapCountByPage)
		if err != nil {
			pagesCount = 0
			return err
		}

		ids, err = database.SwapPage(db, model.SwapID(startID), DefaulSwapCountByPage)
		if err != nil {
			ids = nil
			return err
		}
		for _, id := range ids {
			var info SwapInfo

			swap, err := database.GetSwap(db, id)
			if err != nil {
				ids = nil
				return err
			}

			swapInfo, err := database.GetSwapInfoBySwapID(db, id)
			if err != nil {
				ids = nil
				return err
			}

			info.Timestamp = makeTimestampMillis(swap.Timestamp)
			info.Status = string(swapInfo.Status)

			infos[id] = info
		}

		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("SwapPage failed")
		return sessions.ErrInternalError
	}

	var next string
	if len(ids) > 0 {
		nextID := int(ids[len(ids)-1]) + 1
		secureID, err := sID.ToSecureID("swap", secureid.Value(uint64(nextID)))
		if err != nil {
			return err
		}
		next = sID.ToString(secureID)
	}

	var swaps []SwapInfo
	for _, id := range ids {
		secureID, err := sID.ToSecureID("swap", secureid.Value(uint64(id)))
		if err != nil {
			return err
		}

		var swapInfo SwapInfo
		if info, ok := infos[id]; ok {
			swapInfo = info
		}
		swapInfo.SwapID = sID.ToString(secureID)

		swaps = append(swaps, swapInfo)
	}

	*reply = SwapListResponse{
		RequestPaging: RequestPaging{
			Page:         request.Page,
			PageCount:    pagesCount,
			CountPerPage: DefaulSwapCountByPage,
			Start:        request.Start,
			Next:         next,
		},
		Swaps: swaps,
	}

	return nil
}

// SwapDetailRequest holds args for swapdetail requests
type SwapDetailRequest struct {
	sessions.SessionArgs
	SwapID string `json:"swapId"`
}

// SwapDetailResponse holds response for swapdetail request
type SwapDetailResponse struct {
	SwapID     string `json:"swapId"`
	Timestamp  int64  `json:"timestamp"`
	ValidUntil int64  `json:"validUntil"`
	Status     string `json:"status"`
}

func (p *DashboardService) SwapDetail(r *http.Request, request *SwapDetailRequest, reply *SwapDetailResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.SwapDetail")
	log = networking.GetServiceRequestLog(log, r, "Dashboard", "SwapList")

	// Get userID from session
	request.SessionID = sessions.GetSessionCookie(r)
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

	swapID, err := sID.FromSecureID("swap", sID.Parse(request.SwapID))
	if err != nil {
		log.WithError(err).
			WithField("SwapID", request.SwapID).
			Error("swapID FromSecureID failed")
		return sessions.ErrInternalError
	}

	var swap model.Swap
	var swapStatus model.SwapStatus
	err = db.Transaction(func(db bank.Database) error {
		var err error

		swap, err = database.GetSwap(db, model.SwapID(swapID))
		if err != nil {
			return err
		}

		swapInfo, err := database.GetSwapInfoBySwapID(db, swap.ID)
		if err != nil {
			return err
		}
		swapStatus = swapInfo.Status

		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("SwapDetail failed")
		return sessions.ErrInternalError
	}

	secureID, err := sID.ToSecureID("swap", secureid.Value(uint64(swap.ID)))
	if err != nil {
		return err
	}

	*reply = SwapDetailResponse{
		SwapID:     sID.ToString(secureID),
		Timestamp:  makeTimestampMillis(swap.Timestamp),
		ValidUntil: makeTimestampMillis(swap.ValidUntil),
		Status:     string(swapStatus),
	}

	return nil
}
