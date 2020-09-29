// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"net/http"

	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/condensat/bank-core/logger"

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

// StatusResponse holds args for string requests
type StatusResponse struct {
	Bank BankStatus `json:"bank"`
	Logs LogStatus  `json:"logs"`
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

	bankStatus, logStatus, err := FetchDashboardStatus(ctx)
	if err != nil {
		return apiservice.ErrServiceInternalError
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
