// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"
)

// UserListRequest holds args for userlist requests
type UserListRequest struct {
	RequestPaging
	apiservice.SessionArgs
}

type UserInfo struct {
	UserID string `json:"userId"`
}

// UserListResponse holds response for userlist request
type UserListResponse struct {
	RequestPaging
	Users []UserInfo `json:"users"`
}

func (p *DashboardService) UserList(r *http.Request, request *UserListRequest, reply *UserListResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.UserList")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "UserList")

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

	*reply = UserListResponse{
		RequestPaging: RequestPaging{Page: request.Page, PageCount: 42},
	}

	return nil
}

type UsersStatus struct {
	Count     int `json:"count"`
	Connected int `json:"connected"`
}

func FetchUserStatus(ctx context.Context) (UsersStatus, error) {
	db := appcontext.Database(ctx)
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return UsersStatus{}, err
	}

	userCount, err := database.UserCount(db)
	if err != nil {
		return UsersStatus{}, err
	}
	sessionCount, err := session.Count(ctx)
	if err != nil {
		return UsersStatus{}, err
	}

	return UsersStatus{
		Count:     userCount,
		Connected: sessionCount,
	}, nil
}
