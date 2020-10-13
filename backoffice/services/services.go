// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"

	"github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/networking"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
)

func RegisterServices(ctx context.Context, mux *mux.Router, corsAllowedOrigins []string) {
	corsHandler := networking.CreateCorsOptions(corsAllowedOrigins)

	mux.Handle("/api/v1/dashboard", corsHandler.Handler(NewDashboardHandler(ctx)))
}

func NewDashboardHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := services.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(DashboardService), "dashboard")
	if err != nil {
		panic(err)
	}

	return server
}
