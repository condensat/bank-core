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

	"github.com/condensat/bank-core/monitor"
	"github.com/condensat/bank-core/monitor/database/model"

	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/networking/sessions"

	"github.com/sirupsen/logrus"
)

// StackService receiver
type StackService int

// StackInfoRequest holds args for info requests
type StackInfoRequest struct {
	sessions.SessionArgs
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

// StackInfoResponse holds args for info response
type StackInfoResponse struct {
	Services []ServiceInfo `json:"services"`
}

// StackHistoryRequest holds args for history requests
type StackHistoryRequest struct {
	sessions.SessionArgs
	AppName string `json:"appName"`
	From    int64  `json:"from"`
	To      int64  `json:"to"`
	Step    uint64 `json:"step"`
	Round   uint64 `json:"round"`
}

// ServiceHistory holds service history
type ServiceHistory struct {
	ServiceCount uint64    `json:"serviceCount"`
	Timestamp    []int64   `json:"timestamp"`
	Memory       []uint64  `json:"memory"`
	MemoryMax    []uint64  `json:"memoryMax"`
	ThreadCount  []uint64  `json:"threadCount"`
	CPU          []float64 `json:"cpu"`
}

// StackServiceHistoryResponse holds args for history resonse
type StackServiceHistoryResponse struct {
	AppName string         `json:"appName"`
	History ServiceHistory `json:"history"`
}

// ServiceList operation return the list of active services
func (p *StackService) ServiceList(r *http.Request, request *StackInfoRequest, reply *StackInfoResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.StackService.ServiceList")
	log = networking.GetServiceRequestLog(log, r, "Stack", "ServiceList")

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

func StackListServiceRequest(ctx context.Context) (monitor.StackListService, error) {
	log := logger.Logger(ctx).WithField("Method", "StackService.StackListServiceRequest")
	nats := appcontext.Messaging(ctx)
	var result monitor.StackListService

	message := bank.ToMessage(appcontext.AppName(ctx), &monitor.StackListService{
		Since: time.Hour,
	})
	response, err := nats.Request(ctx, monitor.StackListSubject, message)
	if err != nil {
		log.WithError(err).
			WithField("Subject", monitor.StackListSubject).
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

// ServiceHistory operation return the service history
func (p *StackService) ServiceHistory(r *http.Request, request *StackHistoryRequest, reply *StackServiceHistoryResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.StackService.ServiceHistory")
	log = networking.GetServiceRequestLog(log, r, "Stack", "ServiceHistory")

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

	// Request Service History
	serviceHistory, err := StackServiceHistoryRequest(ctx, request)
	if err != nil {
		log.WithError(err).
			Error("StackListRequest Failed")
		return ErrServiceInternalError
	}

	// Reply
	*reply = StackServiceHistoryResponse{
		AppName: request.AppName,
		History: serviceHistory,
	}

	log.WithFields(logrus.Fields{
		"AppName": reply.AppName,
		"Count":   reply.History.ServiceCount,
	}).Debug("Stack Service History")

	return nil
}

func StackServiceHistoryRequest(ctx context.Context, request *StackHistoryRequest) (ServiceHistory, error) {
	log := logger.Logger(ctx).WithField("Method", "StackService.StackServiceHistoryRequest")
	nats := appcontext.Messaging(ctx)
	var result ServiceHistory

	message := bank.ToMessage(appcontext.AppName(ctx), &monitor.StackServiceHistory{
		AppName: request.AppName,
		From:    time.Unix(request.From, 0),
		To:      time.Unix(request.To, 0),
		Step:    time.Duration(request.Step) * time.Second,
		Round:   time.Duration(request.Round) * time.Second,
	})
	response, err := nats.Request(ctx, monitor.StackServiceHistorySubject, message)
	if err != nil {
		log.WithError(err).
			WithField("Subject", monitor.StackServiceHistorySubject).
			Error("nats.Request Failed")
		return result, ErrServiceInternalError
	}

	var serviceHistory monitor.StackServiceHistory
	err = bank.FromMessage(response, &serviceHistory)
	if err != nil {
		log.WithError(err).
			Error("Message data is not StackListService")
		return result, ErrServiceInternalError
	}

	// initialize historyMap
	var historyMap = make(map[time.Time]*ServiceHistory)
	for _, pi := range serviceHistory.History {
		// create map entry with timestamp
		if _, ok := historyMap[pi.Timestamp]; !ok {
			historyMap[pi.Timestamp] = &ServiceHistory{}
		}
	}

	// append service info with same timestamp (tick)
	for _, pi := range serviceHistory.History {
		// request for appName only, should not append
		if pi.AppName != request.AppName {
			continue
		}
		// append process info to current tick
		tick := historyMap[pi.Timestamp]
		appendInfo(tick, &pi)
	}

	// map to slice
	var ticks []*ServiceHistory
	for _, tick := range historyMap {
		ticks = append(ticks, tick)
	}

	// aggregate services infos
	for _, tick := range ticks {
		result.ServiceCount = tick.ServiceCount
		if len(tick.Timestamp) > 0 {
			result.Timestamp = append(result.Timestamp, tick.Timestamp[0])
		}
		result.Memory = append(result.Memory, cumulateUint(tick.Memory))
		result.MemoryMax = append(result.MemoryMax, cumulateUint(tick.MemoryMax))
		result.ThreadCount = append(result.ThreadCount, cumulateUint(tick.ThreadCount))
		result.CPU = append(result.CPU, cumulateFloat(tick.CPU))
	}

	// order history
	sort.Slice(result.Timestamp, func(i, j int) bool {
		return result.Timestamp[i] < result.Timestamp[j]
	})

	return result, nil
}

func appendInfo(tick *ServiceHistory, pi *model.ProcessInfo) {
	tick.ServiceCount++
	tick.Timestamp = append(tick.Timestamp, pi.Timestamp.UnixNano()/int64(time.Second))
	tick.Memory = append(tick.Memory, pi.MemAlloc)
	tick.MemoryMax = append(tick.MemoryMax, pi.MemTotalAlloc)
	tick.ThreadCount = append(tick.ThreadCount, pi.NumGoroutine)
	tick.CPU = append(tick.CPU, pi.CPUUsage)
}

func cumulateUint(values []uint64) uint64 {
	var ret uint64
	for _, val := range values {
		ret += val
	}
	return ret
}

func cumulateFloat(values []float64) float64 {
	var ret float64
	for _, val := range values {
		ret += val
	}
	return ret
}
