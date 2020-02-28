// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/monitor/common"
	"github.com/condensat/bank-core/monitor/messaging"

	coreService "github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"

	"github.com/sirupsen/logrus"
)

// StackService receiver
type StackService int

// StackInfoRequest holds args for start requests
type StackInfoRequest struct {
	coreService.SessionArgs
}

// ServiceInfo holds service status
type ServiceInfo struct {
	AppName      string  `json:"appName"`
	ServiceCount uint64  `json:"serviceCount"`
	Memory       uint64  `json:"memory"`
	MemoryMax    uint64  `json:"memoryMax"`
	ThreadCount  uint64  `json:"threadCount"`
	CPU          float64 `json:"cpu"`
}

// StackInfoResponse holds args for start requests
type StackInfoResponse struct {
	Services []ServiceInfo `json:"services"`
}

// ServiceList operation return the list of active services
func (p *StackService) ServiceList(r *http.Request, request *StackInfoRequest, reply *StackInfoResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.StackService.ServiceList")
	log = coreService.GetServiceRequestLog(log, r, "Stack", "ServiceList")

	verified, err := verifySessionId(ctx, sessions.SessionID(request.SessionID))
	if err != nil {
		log.WithError(err).
			Error("verifySessionId Failed")
		return ErrServiceInternalError
	}

	if !verified {
		log.Error("Invalid sessionId")
		return sessions.ErrInvalidSessionID
	}

	// Request Service List
	listService, err := StackListServiceRequest(ctx)
	if err != nil {
		log.WithError(err).
			Error("StackListRequest Failed")
		return ErrServiceInternalError
	}

	// Reply
	reply.Services = make([]ServiceInfo, len(listService.Services))
	for i, service := range listService.Services {
		reply.Services[i].AppName = service
	}

	for _, info := range listService.ProcessInfo {
		for i := range reply.Services {
			if reply.Services[i].AppName != info.AppName {
				continue
			}

			reply.Services[i].ServiceCount++
			reply.Services[i].Memory += info.MemAlloc
			reply.Services[i].MemoryMax += info.MemSys
			reply.Services[i].ThreadCount += info.NumGoroutine
			reply.Services[i].CPU += info.CPUUsage
		}
	}

	log.WithFields(logrus.Fields{
		"Services": reply.Services,
	}).Debug("Stack Services")

	return nil
}

func StackListServiceRequest(ctx context.Context) (common.StackListService, error) {
	log := logger.Logger(ctx).WithField("Method", "StackService.StackListServiceRequest")
	nats := appcontext.Messaging(ctx)
	var result common.StackListService

	message := bank.ToMessage(appcontext.AppName(ctx), &common.StackListService{
		Since: time.Hour,
	})
	response, err := nats.Request(ctx, messaging.StackListSubject, message)
	if err != nil {
		log.WithError(err).
			WithField("Subject", messaging.StackListSubject).
			Error("nats.Request Failed")
		return result, ErrServiceInternalError
	}

	err = bank.FromMessage(response, &result)
	if err != nil {
		log.WithError(err).
			Error("Message data is not StackListService")
		return result, ErrServiceInternalError
	}

	sort.Slice(result.Services, func(i, j int) bool {
		return result.Services[i] < result.Services[j]
	})

	return result, nil
}
