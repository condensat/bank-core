// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/condensat/bank-core/api"
	coreService "github.com/condensat/bank-core/api/services"

	"github.com/condensat/bank-core/monitor/services"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type StackMonitor int

func (p *StackMonitor) Run(ctx context.Context, port int, corsAllowedOrigins []string) {
	log := logger.Logger(ctx).WithField("Method", "monitor.StackMonitor.Run")
	muxer := http.NewServeMux()

	services.RegisterServices(ctx, muxer, corsAllowedOrigins)

	handler := negroni.New(&negroni.Recovery{})
	handler.Use(coreService.StatsMiddleware)
	handler.UseFunc(api.MiddlewarePeerRateLimiter)
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
	}).Info("Stack Monintor Service started")

	<-ctx.Done()
}

// AddWorkerHeader - adds header of which node actually processed request
func AddWorkerHeader(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker", utils.Hostname())
	next(rw, r)
}

// AddWorkerVersion - adds header of which version is installed
func AddWorkerVersion(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker-Version", coreService.Version)
	next(rw, r)
}
