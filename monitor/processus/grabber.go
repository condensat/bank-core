// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package processus

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/monitor"
	"github.com/condensat/bank-core/utils"
)

type Grabber struct {
	appName   string
	interval  time.Duration
	messaging bank.Messaging
}

func NewGrabber(ctx context.Context, interval time.Duration) *Grabber {
	return &Grabber{
		appName:   appcontext.AppName(ctx),
		interval:  interval,
		messaging: appcontext.Messaging(ctx),
	}
}

func (p *Grabber) Run(ctx context.Context, numWorkers int) {
	log := logger.Logger(ctx).WithField("Method", "processus.Grabber.Run")

	for {
		select {
		case <-time.After(p.interval):

			processInfo := processInfo(p.appName)
			err := p.sendProcessInfo(ctx, &processInfo)
			if err != nil {
				log.WithError(err).Error("Failed to sendProcessInfo")
				continue
			}
			log.Trace("Grab processInfo")

		case <-ctx.Done():
			log.Info("Process Grabber done.")
			return
		}
	}
}

func processInfo(appName string) monitor.ProcessInfo {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return monitor.ProcessInfo{
		Timestamp: time.Now().UTC().Truncate(time.Second),
		AppName:   appName,
		Hostname:  utils.Hostname(),
		PID:       os.Getpid(),

		MemAlloc:      mem.Alloc,
		MemTotalAlloc: mem.TotalAlloc,
		MemSys:        mem.Sys,
		MemLookups:    mem.Lookups,

		NumCPU:       uint64(runtime.NumCPU()),
		NumGoroutine: uint64(runtime.NumGoroutine()),
		NumCgoCall:   uint64(runtime.NumCgoCall()),
	}
}

func (p *Grabber) sendProcessInfo(ctx context.Context, processInfo *monitor.ProcessInfo) error {
	request := bank.ToMessage(p.appName, processInfo)
	return p.messaging.Publish(ctx, "Condensat.Monitor.Inbound", request)
}
