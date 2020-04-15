// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oauth

import (
	"context"
	"crypto/sha512"
	"errors"
	"net/http"
	"os"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/joho/godotenv"
)

var (
	ErrInvalidOAuthKeys   = errors.New("Invalid OAuth keys file")
	ErrInvalidOAuthDomain = errors.New("Invalid OAuth Domain")
)

type Options struct {
	Keys   string
	Domain string
}

func Init(options Options) error {
	err := godotenv.Overload(options.Keys)
	if err != nil {
		return ErrInvalidOAuthKeys
	}
	if len(options.Domain) == 0 {
		return ErrInvalidOAuthDomain
	}

	key := sha512.Sum512([]byte(os.Getenv("BANK_OAUTH_SESSION_SECRET")))
	cookieStore := sessions.NewCookieStore(key[:32], key[32:])
	cookieStore.Options.Path = "/api/v1/auth"
	cookieStore.Options.Domain = options.Domain
	cookieStore.Options.Secure = true
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = cookieStore

	goth.UseProviders(
		github.New(os.Getenv("OAUTH_GITHUB_KEY"), os.Getenv("OAUTH_GITHUB_SECRET"), os.Getenv("OAUTH_GITHUB_CALLBACK")),
	)

	return nil
}

// Register handlers for OAuth providers
func RegisterHandlers(ctx context.Context, server *mux.Router) {
	server.HandleFunc("/api/v1/auth/{provider}", AuthHandler)
	server.HandleFunc("/api/v1/auth/{provider}/callback", AuthCallbackHandler)
	server.HandleFunc("/api/v1/auth/{provider}/logout", AuthLogoutHandler)
}

// AuthHandler reuse oauth session or open a new one
var AuthHandler = func(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := logger.Logger(ctx).WithField("Method", "oauth.AuthHandler")

	// try to get the user without re-authenticating
	if user, err := gothic.CompleteUserAuth(res, req); err == nil {
		err = UpdateUserSession(ctx, req, res, user)
		if err != nil {
			log.WithError(err).Errorf("UpdateUserSession failed")
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

// AuthCallbackHandler finalize oauth authentification
var AuthCallbackHandler = func(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := logger.Logger(ctx).WithField("Method", "oauth.AuthCallbackHandler")

	webAppUrl := appcontext.WebAppURL(ctx)

	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		log.WithError(err).Errorf("CompleteUserAuth failed")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = UpdateUserSession(ctx, req, res, user)
	if err != nil {
		log.WithError(err).Errorf("UpdateUserSession failed")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(res, req, webAppUrl, http.StatusFound)
}

// AuthLogoutHandler close oauth session
var AuthLogoutHandler = func(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := logger.Logger(ctx).WithField("Method", "oauth.AuthLogoutHandler")
	webAppUrl := appcontext.WebAppURL(ctx)

	err := RemoveSession(ctx, req)
	if err != nil {
		log.WithError(err).
			Warning("RemoveSession failed")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = gothic.Logout(res, req)
	if err != nil {
		log.WithError(err).
			Warning("OAuth Logout failed")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(res, req, webAppUrl, http.StatusFound)
}
