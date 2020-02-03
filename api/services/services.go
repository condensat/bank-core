// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/appcontext"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
)

var (
	ErrServiceInternalError = errors.New("Service Internal Error")
)

func RegisterServices(ctx context.Context, mux *http.ServeMux) {
	mux.Handle("/api/v1/session", NewSessionHandler())
}

func NewSessionHandler() http.Handler {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterService(new(SessionService), "session")

	return server
}

func fromContext(ctx context.Context) (db bank.Database, session *sessions.Session, err error) {
	db = appcontext.Database(ctx)
	if ctxSession, ok := ctx.Value(sessions.KeySessions).(*sessions.Session); ok {
		session = ctxSession
	}
	if db == nil || session == nil {
		db = nil
		session = nil
		err = ErrServiceInternalError
		return
	}

	return
}
