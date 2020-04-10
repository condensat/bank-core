// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/condensat/bank-core/api/sessions"
)

func getSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func UserIDFromSession(ctx context.Context, req *http.Request) (uint64, error) {
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return 0, err
	}

	sessionID := sessions.SessionID(getSessionCookie(req))
	userID := session.UserSession(ctx, sessionID)
	if !sessions.IsUserValid(userID) {
		return 0, errors.New("Invalid UserID")
	}

	return userID, nil
}

func RemoveSession(ctx context.Context, req *http.Request) error {
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return err
	}

	sessionID := sessions.SessionID(getSessionCookie(req))

	return session.InvalidateSession(ctx, sessionID)
}
