// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rate

import (
	"context"
	"fmt"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
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
	log := logger.Logger(ctx).WithField("Method", "RateGrabber.Run")
	appID = appcontext.SecretOrPassword(appID)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("RateGrabber started")

	// get currency from database and store to redis
	currencyRates, err := database.GetLastCurencyRates(ctx)
	if err != nil {
		log.WithError(err).
			Warning("No currencies found in database")
	}

	UpdateRedisRate(ctx, currencyRates)

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
	log := logger.Logger(ctx).WithField("Method", "RateGrabber.scheduledGrabber")

	interval, delay = checkParams(interval, delay)

	log = log.WithFields(logrus.Fields{
		"Interval": fmt.Sprintf("%s", interval),
		"Delay":    fmt.Sprintf("%s", delay),
	})

	log.Info("Start grabber Scheduler")

	for epoch := range utils.Scheduler(ctx, interval, delay) {
		currencyRates, err := FetchLatestRates(ctx, appID)
		if err != nil {
			log.WithError(err).
				Error("Failed to FetchLatestRates")
			continue
		}

		if len(currencyRates) == 0 {
			log.
				Warning("FetchLatestRates returns empty currency rates")
			continue
		}

		err = database.AppendCurencyRates(ctx, currencyRates)
		if err != nil {
			log.WithError(err).
				Error("Failed to addCurencyRates")
			continue
		}

		log.WithFields(logrus.Fields{
			"Epoch": epoch.Truncate(time.Millisecond),
			"Count": len(currencyRates),
		}).Debug("CurrencyRates stored")

		UpdateRedisRate(ctx, currencyRates)
	}
}
