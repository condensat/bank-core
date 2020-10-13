// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/condensat/bank-core/database/model"

	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/networking/sessions"

	"github.com/condensat/bank-core/logger"
)

var (
	ErrPermissionDenied = errors.New("Permission Denied")
)

type DashboardService int

// StatusRequest holds args for status requests
type StatusRequest struct {
	sessions.SessionArgs
}

// StatusResponse holds args for string requests
type StatusResponse struct {
	Bank BankStatus `json:"bank"`
	Logs LogStatus  `json:"logs"`
}

func (p *DashboardService) Status(r *http.Request, request *StatusRequest, reply *StatusResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.Status")
	log = networking.GetServiceRequestLog(log, r, "Dashboard", "Status")

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

	bankStatus, logStatus, err := FetchDashboardStatus(ctx)
	if err != nil {
		return sessions.ErrInternalError
	}
	*reply = StatusResponse{
		Bank: bankStatus,
		Logs: logStatus,
	}

	return nil
}

func FetchDashboardStatus(ctx context.Context) (BankStatus, LogStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "DashboardService.FetchDashboardStatus")

	bankStatus, err := FetchBankStatus(ctx)
	if err != nil {
		log.WithError(err).
			Error("FetchBankStatus failed")
		return BankStatus{}, LogStatus{}, err
	}

	logStatus, err := FetchLogStatus(ctx)
	if err != nil {
		log.WithError(err).
			Error("FetchLogStatus failed")
		return BankStatus{}, LogStatus{}, err
	}

	return bankStatus, logStatus, nil
}
