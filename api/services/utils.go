// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"
	"time"

	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

func makeTimestampMillis(ts time.Time) int64 {
	return ts.UnixNano() / int64(time.Millisecond)
}

func fromTimestampMillis(timestamp int64) time.Time {
	return time.Unix(0, int64(timestamp)*int64(time.Millisecond)).UTC()
}

func AppendRequestLog(log *logrus.Entry, r *http.Request) *logrus.Entry {
	return log.WithFields(logrus.Fields{
		"UserAgent": r.UserAgent(),
		"IP":        r.RemoteAddr,
		"URI":       r.RequestURI,
	})
}

func GetRequestLog(ctx context.Context, r *http.Request) *logrus.Entry {
	return AppendRequestLog(logger.Logger(ctx), r)
}

func GetServiceRequestLog(log *logrus.Entry, r *http.Request, service, operation string) *logrus.Entry {
	log = AppendRequestLog(log, r)

	// Optionals
	if len(service) > 0 {
		log = log.WithField("Service", service)
	}
	if len(service) > 0 {
		log = log.WithField("Operation", operation)
	}

	return log
}
