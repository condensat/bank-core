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
	"github.com/condensat/bank-core/logger"

	"github.com/gorilla/rpc/v2"
)

var (
	ErrServiceInternalError = errors.New("Service Internal Error")
)

func RegisterMessageHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "RegisterMessageHandlers")

	nats := appcontext.Messaging(ctx)
	nats.SubscribeWorkers(ctx, VerifySessionSubject, 4, sessions.VerifySession)

	log.Debug("MessageHandlers registered")
}

func RegisterServices(ctx context.Context, mux *http.ServeMux, corsAllowedOrigins []string) {
	corsHandler := CreateCorsOptions(corsAllowedOrigins)

	mux.Handle("/api/v1/session", corsHandler.Handler(NewSessionHandler(ctx)))
	mux.Handle("/api/v1/user", corsHandler.Handler(NewUserHandler(ctx)))
}

func NewSessionHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(SessionService), "session")
	if err != nil {
		panic(err)
	}

	return server
}

func NewUserHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(UserService), "user")
	if err != nil {
		panic(err)
	}

	return server
}

func ContextValues(ctx context.Context) (bank.Database, *sessions.Session, error) {
	db := appcontext.Database(ctx)
	session, err := sessions.ContextSession(ctx)
	if db == nil || session == nil {
		err = ErrServiceInternalError
	}

	if err != nil {
		return nil, nil, err
	}

	return db, session, nil
}
