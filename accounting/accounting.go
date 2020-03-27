// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"errors"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/monitor/messaging"
	"github.com/condensat/bank-core/utils"

	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrAddProcessInfo = errors.New("AddProcessInfo")
	ErrInternalError  = errors.New("InternalError")
)

type Accounting int

func (p *Accounting) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.Run")

	p.registerHandlers(RedisMutexContext(ctx))

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Accounting Service started")

	<-ctx.Done()
}

func (p *Accounting) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)
	nats.SubscribeWorkers(ctx, messaging.InboundSubject, 8, p.onUserAccounts)
	nats.SubscribeWorkers(ctx, messaging.InboundSubject, 8, p.onAccountHistory)

	log.Debug("Bank Accounting registered")
}

func (p *Accounting) onUserAccounts(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.onUserAccounts")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var req UserAccounts
	err := bank.FromMessage(message, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not UserAccounts")
		return nil, ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"UserID": req.UserID,
	})
	accounts, err := ListUserAccounts(ctx, req.UserID)
	if err != nil {
		log.WithError(err).Errorf("Failed to ListUserAccounts")
	}

	// create response
	resp := UserAccounts{
		UserID: req.UserID,

		Accounts: accounts[:],
	}

	return bank.ToMessage(appcontext.AppName(ctx), &resp), nil
}

func (p *Accounting) onAccountHistory(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.onAccountHistory")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var req AccountHistory
	err := bank.FromMessage(message, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not AccountHistory")
		return nil, ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"AccountID": req.AccountID,
	})

	history, err := GetAccountHistory(ctx, req.AccountID, req.From, req.To)
	if err != nil {
		log.WithError(err).Errorf("Failed to AccountHistory")
	}

	// create response
	resp := AccountHistory{
		AccountID: req.AccountID,
		From:      req.From,
		To:        req.To,

		History: history,
	}

	return bank.ToMessage(appcontext.AppName(ctx), &resp), nil
}
