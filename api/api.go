// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/condensat/bank-core/api/services"
	"github.com/condensat/bank-core/logger"

	"github.com/urfave/negroni"
)

type Api int

func (p *Api) Run(ctx context.Context, port int) {
	muxer := http.NewServeMux()

	handler := negroni.New(&negroni.Recovery{})
	handler.Use(services.StatsMiddleware)
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
			logger.Logger(ctx).
				WithError(err).
				Info("Http server exited")
		}
	}()

	logger.Logger(ctx).
		WithField("Port", port).
		Info("Api Service started")

	<-ctx.Done()
}

// AddWorkerHeader - adds header of which node actually processed request
func AddWorkerHeader(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker", getHost())
	next(rw, r)
}

// AddWorkerVersion - adds header of which version is installed
func AddWorkerVersion(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker-Version", services.Version)
	next(rw, r)
}

func getHost() string {
	var err error
	host, err := os.Hostname()
	if err != nil {
		host = "Unknown"
	}
	return host
}
