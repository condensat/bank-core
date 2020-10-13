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
	DefaultUserCountByPage = 10
)

// UserListRequest holds args for userlist requests
type UserListRequest struct {
	sessions.SessionArgs
	RequestPaging
}

type UserInfo struct {
	UserID string   `json:"userId"`
	Name   string   `json:"name,omitempty"`
	Email  string   `json:"email,omitempty"`
	Roles  []string `json:"roles,omitempty"`
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
	log = networking.GetServiceRequestLog(log, r, "Dashboard", "UserList")

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
		startID, err = sID.FromSecureID("user", sID.Parse(request.Start))
		if err != nil {
			log.WithError(err).
				WithField("Start", request.Start).
				Error("startID FromSecureID failed")
			return ErrPermissionDenied
		}
	}
	var pagesCount int
	var userPage []model.User
	err = db.Transaction(func(db bank.Database) error {
		var err error
		pagesCount, err = database.UserPagingCount(db, DefaultUserCountByPage)
		if err != nil {
			pagesCount = 0
			return err
		}

		userPage, err = database.UserPage(db, model.UserID(startID), DefaultUserCountByPage)
		if err != nil {
			userPage = nil
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("UserPaging failed")
		return sessions.ErrInternalError
	}

	var next string
	if len(userPage) > 0 {
		nextID := int(userPage[len(userPage)-1].ID) + 1
		secureID, err := sID.ToSecureID("user", secureid.Value(uint64(nextID)))
		if err != nil {
			return err
		}
		next = sID.ToString(secureID)
	}

	var users []UserInfo
	for _, user := range userPage {
		secureID, err := sID.ToSecureID("user", secureid.Value(uint64(user.ID)))
		if err != nil {
			return err
		}

		users = append(users, UserInfo{
			UserID: sID.ToString(secureID),
			Email:  string(user.Email),
		})
	}

	*reply = UserListResponse{
		RequestPaging: RequestPaging{
			Page:         request.Page,
			PageCount:    pagesCount,
			CountPerPage: DefaultUserCountByPage,
			Start:        request.Start,
			Next:         next,
		},
		Users: users[:],
	}

	return nil
}

// UserListRequest holds args for userdetail requests
type UserDetailRequest struct {
	sessions.SessionArgs
	UserID string `json:"userId"`
}

// UserDetailResponse holds response for userdetail request
type UserDetailResponse struct {
	Info UserInfo `json:"userInfo"`
}

func (p *DashboardService) UserDetail(r *http.Request, request *UserDetailRequest, reply *UserDetailResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.UserDetail")
	log = networking.GetServiceRequestLog(log, r, "Dashboard", "UserDetail")

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

	userID, err := sID.FromSecureID("user", sID.Parse(request.UserID))
	if err != nil {
		log.WithError(err).
			WithField("UserID", request.UserID).
			Error("userID FromSecureID failed")
		return ErrPermissionDenied
	}

	var user model.User
	var roles []string
	err = db.Transaction(func(db bank.Database) error {
		var err error

		user, err = database.FindUserById(db, model.UserID(userID))
		if err != nil {
			return err
		}

		roleNames, err := database.UserRoles(db, user.ID)
		if err != nil {
			return err
		}
		for _, roleName := range roleNames {
			roles = append(roles, string(roleName))

		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("UserDetails failed")
		return sessions.ErrInternalError
	}

	*reply = UserDetailResponse{
		Info: UserInfo{
			UserID: request.UserID,
			Name:   string(user.Name),
			Roles:  roles,
		},
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
