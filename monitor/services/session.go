// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/networking/sessions"
)

func verifySessionId(ctx context.Context, sessionID sessions.SessionID) (bool, error) {
	log := logger.Logger(ctx).WithField("Method", "verifySessionId")
	nats := messaging.FromContext(ctx)

	message := messaging.ToMessage(appcontext.AppName(ctx), &sessions.SessionInfo{
		SessionID: sessionID,
	})
	if message == nil {
		log.Error("messaging.ToMessage failed")
		return false, ErrServiceInternalError
	}

	response, err := nats.Request(ctx, ApiVerifySessionSubject, message)
	if err != nil {
		log.WithError(err).
			WithField("Subject", ApiVerifySessionSubject).
			Error("nats.Request Failed")
		return false, ErrServiceInternalError
	}

	var si sessions.SessionInfo
	err = messaging.FromMessage(response, &si)
	if err != nil {
		log.WithError(err).
			Error("Message data is not SessionInfo")
		return false, ErrServiceInternalError
	}

	return sessionID == si.SessionID &&
		sessions.IsSessionValid(si.SessionID) &&
		sessions.IsUserValid(si.UserID) &&
		!si.Expired(), nil
}
