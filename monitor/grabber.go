// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/utils"

	"github.com/condensat/bank-core/monitor/database"
	"github.com/condensat/bank-core/monitor/database/model"

	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrAddProcessInfo = errors.New("AddProcessInfo")
	ErrInternalError  = errors.New("InternalError")
)

type Grabber int

func (p *Grabber) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "grabber.Run")

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Grabber Service started")

	<-ctx.Done()
}

func (p *Grabber) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "grabber.RegisterHandlers")

	nats := appcontext.Messaging(ctx)
	nats.SubscribeWorkers(ctx, InboundSubject, 4, p.onProcessInfo)
	nats.SubscribeWorkers(ctx, StackListSubject, 4, p.onStackList)
	nats.SubscribeWorkers(ctx, StackServiceHistorySubject, 4, p.onStackServiceHistory)

	log.Debug("Monitor Grabber registered")
}

func (p *Grabber) onProcessInfo(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "grabber.onProcessInfo")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	db := appcontext.Database(ctx)

	var req model.ProcessInfo
	err := bank.FromMessage(message, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not model.ProcessInfo")
		return nil, ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"AppName":  req.AppName,
		"Hostname": req.Hostname,
	})

	err = database.AddProcessInfo(db, &req)
	if err != nil {
		log.WithError(err).Error("Failed to AddProcessInfo")
		return nil, ErrAddProcessInfo
	}

	return nil, nil
}

func (p *Grabber) onStackList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "grabber.onStackList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var req StackListService
	err := bank.FromMessage(message, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not StackListService")
		return nil, ErrInternalError
	}

	db := appcontext.Database(ctx)

	processInfo, err := database.LastServicesStatus(db)
	if err != nil {
		log.WithError(err).Error("LastServicesStatus failed")
		return nil, ErrInternalError
	}

	// find unique names
	var serviceMap = make(map[string]string)
	for _, pi := range processInfo {
		if _, ok := serviceMap[pi.AppName]; ok {
			continue
		}
		serviceMap[pi.AppName] = pi.AppName
	}

	// get unique names
	var services = make([]string, 0, len(serviceMap))
	for appName := range serviceMap {
		services = append(services, appName)
	}

	// create response
	resp := StackListService{
		Services:    services[:],
		ProcessInfo: processInfo[:],
	}

	return bank.ToMessage(appcontext.AppName(ctx), &resp), nil
}

func (p *Grabber) onStackServiceHistory(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "grabber.onStackServiceHistory")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var req StackServiceHistory
	err := bank.FromMessage(message, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not StackServiceHistory")
		return nil, ErrInternalError
	}

	db := appcontext.Database(ctx)

	history, err := database.LastServiceHistory(db, req.AppName, req.From, req.To, req.Step, req.Round)
	if err != nil {
		log.WithError(err).Error("LastServiceHistory failed")
		return nil, ErrInternalError
	}

	// create response
	resp := StackServiceHistory{
		AppName: req.AppName,
		History: history,
	}

	return bank.ToMessage(appcontext.AppName(ctx), &resp), nil
}
