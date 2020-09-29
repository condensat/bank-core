// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/condensat/bank-core"
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
	db := appcontext.Database(ctx)
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

	var pagesCount int
	var ids []model.UserID
	const DefaultUserCountByPage = 100
	err = db.Transaction(func(db bank.Database) error {
		var err error
		pagesCount, err = database.UserPagingCount(db, DefaultUserCountByPage)
		if err != nil {
			pagesCount = 0
			return err
		}

		ids, err = database.UserPage(db, request.Page, DefaultUserCountByPage)
		if err != nil {
			ids = nil
			return err
		}

		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("UserPaging failed")
		return apiservice.ErrServiceInternalError
	}

	var users []UserInfo
	for _, id := range ids {
		users = append(users, UserInfo{
			UserID: fmt.Sprintf("%d", id),
		})
	}

	*reply = UserListResponse{
		RequestPaging: RequestPaging{Page: request.Page, PageCount: pagesCount},
		Users:         users[:],
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
