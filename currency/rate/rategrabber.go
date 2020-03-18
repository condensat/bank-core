// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rate

import (
	"context"
	"fmt"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

const (
	DefaultInterval time.Duration = time.Hour
	DefaultDelay    time.Duration = 5 * time.Second
)

type RateGrabber int

func (p *RateGrabber) Run(ctx context.Context, appID string, interval time.Duration, delay time.Duration) {
	log := logger.Logger(ctx).WithField("Method", "currency.rate.RateGrabber.Run")
	appID = appcontext.SecretOrPassword(appID)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("RateGrabber started")

	go p.scheduledGrabber(ctx, appID, interval, delay)

	<-ctx.Done()
}

func checkParams(interval time.Duration, delay time.Duration) (time.Duration, time.Duration) {
	if interval < time.Second {
		interval = DefaultInterval
	}
	if delay < 0 {
		delay = DefaultDelay
	}

	return interval, delay
}

func (p *RateGrabber) scheduledGrabber(ctx context.Context, appID string, interval time.Duration, delay time.Duration) {
	log := logger.Logger(ctx).WithField("Method", "currency.rate.RateGrabber.grabRate")

	interval, delay = checkParams(interval, delay)

	log = log.WithFields(logrus.Fields{
		"Interval": fmt.Sprintf("%s", interval),
		"Delay":    fmt.Sprintf("%s", delay),
	})

	log.Info("Start grabber Scheduler")

	for epoch := range utils.Scheduler(ctx, interval, delay) {
		var currencyRates []model.Currency
		log.WithFields(logrus.Fields{
			"Epoch": epoch.Truncate(time.Millisecond),
			"Count": len(currencyRates),
		}).Debug("Got CurrencyRates")
	}
}
