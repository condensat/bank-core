// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"

	"github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/markbates/goth"
)

func getSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func UpdateUserSession(ctx context.Context, req *http.Request, w http.ResponseWriter, user goth.User) error {
	var userID uint64
	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {
		u, err := database.FindUserByEmail(db, model.UserEmail(user.Email))
		if err != nil {
			return err
		}

		// creat user if email does not exists
		if u.ID == 0 {
			u, err = database.FindOrCreateUser(db, model.User{
				Name:  model.UserName(fmt.Sprintf("%s:%s", user.Provider, user.NickName)),
				Email: model.UserEmail(user.Email),
			})
			if err != nil {
				return err
			}
		}

		// store userID for cookie creation
		userID = uint64(u.ID)

		oa, err := database.FindOrCreateOAuth(db, model.OAuth{
			Provider:   user.Provider,
			ProviderID: user.NickName,
			UserID:     u.ID,
		})
		if err != nil {
			return err
		}

		// store oauth data
		data, err := json.Marshal(&user)
		if err != nil {
			return err
		}

		_, err = database.CreateOrUpdateOAuthData(db, model.OAuthData{
			OAuthID: oa.ID,
			Data:    string(data),
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if userID == 0 {
		return sessions.ErrInvalidUserID
	}

	// create session & cookie
	err = services.CreateSessionWithCookie(ctx, req, w, userID)
	if err != nil {
		return err
	}

	return nil
}

func RemoveSession(ctx context.Context, req *http.Request) error {
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return err
	}

	sessionID := sessions.SessionID(getSessionCookie(req))

	return session.InvalidateSession(ctx, sessionID)
}
