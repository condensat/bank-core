// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

func VerifySession(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "services.VerifySession")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	// Retrieve sessions from context
	session, err := ContextSession(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session renew failed")
		return nil, ErrInternalError
	}

	var sessionInfo SessionInfo
	err = bank.FromMessage(message, &sessionInfo)
	if err != nil {
		log.WithError(err).
			Warning("Message data is not SessionInfo")
		return nil, ErrInternalError
	}

	resp := session.sessionInfo(ctx, sessionInfo.SessionID)

	return bank.ToMessage(appcontext.AppName(ctx), &resp), nil
}
