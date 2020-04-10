// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oauth

import (
	"crypto/sha512"
	"errors"
	"net/http"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"

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
	cookieStore := sessions.NewCookieStore(key[:])
	cookieStore.Options.Path = "/api/v1/auth"
	cookieStore.Options.Domain = options.Domain
	cookieStore.Options.Secure = true
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.SameSite = http.SameSiteStrictMode

	gothic.Store = cookieStore

	goth.UseProviders(
		github.New(os.Getenv("OAUTH_GITHUB_KEY"), os.Getenv("OAUTH_GITHUB_SECRET"), os.Getenv("OAUTH_GITHUB_CALLBACK")),
	)

	return nil
}
