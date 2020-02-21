// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type Api int

func (p *Api) Run(ctx context.Context, port int, corsAllowedOrigins []string) {
	log := logger.Logger(ctx).WithField("Method", "api.Api.Run")

	muxer := http.NewServeMux()

	// create session and and to context
	session := sessions.NewSession(ctx)
	ctx = context.WithValue(ctx, sessions.KeySessions, session)

	services.RegisterServices(ctx, muxer, corsAllowedOrigins)

	handler := negroni.New(&negroni.Recovery{})
	handler.Use(services.StatsMiddleware)
	handler.UseFunc(MiddlewarePeerRateLimiter)
	handler.UseFunc(AddWorkerHeader)
	handler.UseFunc(AddWorkerVersion)
	handler.UseHandler(muxer)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
		MaxHeaderBytes: 1 << 16, // 16 KiB
		ConnContext:    func(conCtx context.Context, c net.Conn) context.Context { return ctx },
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.WithError(err).
				Info("Http server exited")
		}
	}()

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
		"Port":     port,
	}).Info("Api Service started")

	<-ctx.Done()
}

// AddWorkerHeader - adds header of which node actually processed request
func AddWorkerHeader(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker", utils.Hostname())
	next(rw, r)
}

// AddWorkerVersion - adds header of which version is installed
func AddWorkerVersion(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker-Version", services.Version)
	next(rw, r)
}
