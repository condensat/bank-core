// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"net/http"

	apiservice "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"

	"github.com/condensat/bank-core/logger"
)

type DashboardService int

// StatusRequest holds args for status requests
type StatusRequest struct {
	apiservice.SessionArgs
}

type UsersStatus struct {
	Count int `json:"count"`
}

// StatusResponse holds args for string requests
type StatusResponse struct {
	Users UsersStatus `json:"users"`
}

func (p *DashboardService) Status(r *http.Request, request *StatusRequest, reply *StatusResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.Status")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "Status")

	userCount, err := database.UserCount(db)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}
	*reply = StatusResponse{
		Users: UsersStatus{
			Count: userCount,
		},
	}

	log.Info("Status")

	return nil
}
