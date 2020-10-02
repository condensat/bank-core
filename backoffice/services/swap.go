// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"

	"github.com/condensat/bank-core"
	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/secureid"
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
	apiservice.SessionArgs
	RequestPaging
}

type SwapInfo struct {
	SwapID string `json:"swapId"`
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
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "SwapList")

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
		startID, err = sID.FromSecureID("swap", sID.Parse(request.Start))
		if err != nil {
			log.WithError(err).
				WithField("Start", request.Start).
				Error("startID FromSecureID failed")
			return apiservice.ErrServiceInternalError
		}
	}
	var pagesCount int
	var ids []model.SwapID
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
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("SwapPage failed")
		return apiservice.ErrServiceInternalError
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

		swaps = append(swaps, SwapInfo{
			SwapID: sID.ToString(secureID),
		})
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
