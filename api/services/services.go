// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/networking/sessions"

	"github.com/gorilla/mux"
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

func RegisterServices(ctx context.Context, mux *mux.Router, corsAllowedOrigins []string) {
	corsHandler := networking.CreateCorsOptions(corsAllowedOrigins)

	mux.Handle("/api/v1/session", corsHandler.Handler(NewSessionHandler(ctx)))
	mux.Handle("/api/v1/user", corsHandler.Handler(NewUserHandler(ctx)))
	mux.Handle("/api/v1/accounting", corsHandler.Handler(NewAccountingHandler(ctx)))
	mux.Handle("/api/v1/wallet", corsHandler.Handler(NewWalletHandler(ctx)))
	mux.Handle("/api/v1/swap", corsHandler.Handler(NewSwapHandler(ctx)))
}

func NewSessionHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := sessions.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(
		NewSessionService(
			func(ctx context.Context, login, password string) (uint64, bool, error) {
				db := appcontext.Database(ctx)
				userID, allowed, err := query.CheckCredential(ctx, db, model.Base58(login), model.Base58(password))
				return uint64(userID), allowed, err
			},
		), "session")
	if err != nil {
		panic(err)
	}

	return server
}

func NewUserHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := sessions.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(UserService), "user")
	if err != nil {
		panic(err)
	}

	return server
}

func NewAccountingHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := sessions.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(AccountingService), "accounting")
	if err != nil {
		panic(err)
	}

	return server
}

func NewWalletHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := sessions.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(WalletService), "wallet")
	if err != nil {
		panic(err)
	}

	return server
}

func NewSwapHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := sessions.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(SwapService), "swap")
	if err != nil {
		panic(err)
	}

	return server
}

func ContextValues(ctx context.Context) (database.Context, *sessions.Session, error) {
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
