// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"

	"github.com/condensat/bank-core/appcontext"

	"github.com/condensat/bank-core/networking/sessions"

	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

func isUserAdmin(ctx context.Context, log *logrus.Entry, sessionID sessions.SessionID) (bool, *logrus.Entry, error) {

	db := appcontext.Database(ctx)
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return false, log, err
	}

	// Get userID from session
	userID := session.UserSession(ctx, sessionID)
	if !sessions.IsUserValid(userID) {
		log.Error("Invalid userSession")
		return false, log, sessions.ErrInvalidSessionID
	}

	log = log.WithFields(logrus.Fields{
		"SessionID": sessionID,
		"UserID":    userID,
	})

	isAdmin, err := query.UserHasRole(db, model.UserID(userID), model.RoleNameAdmin)
	if err != nil {
		log.WithError(err).
			WithField("RoleName", model.RoleNameAdmin).
			Error("UserHasRole failed")
		return false, log, ErrPermissionDenied
	}

	return isAdmin, log, nil
}
