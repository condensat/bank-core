// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rate

import (
	"context"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

type RateGrabber struct {
	appID string
}

func (p *RateGrabber) Run(ctx context.Context, appID string) {
	log := logger.Logger(ctx).WithField("Method", "currency.rate.RateGrabber.Run")
	p.appID = appcontext.SecretOrPassword(appID)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("RateGrabber started")

	<-ctx.Done()
}
