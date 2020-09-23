// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package backoffice

import (
	"context"

	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/utils"

	"github.com/sirupsen/logrus"
)

type BackOffice int

func (p *BackOffice) Run(ctx context.Context, port int, corsAllowedOrigins []string) {
	log := logger.Logger(ctx).WithField("Method", "backoffice.BackOffice.Run")

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
		"Port":     port,
	}).Info("BackOffice Service started")

	<-ctx.Done()
}
