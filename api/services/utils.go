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

func getServiceRequestLog(ctx context.Context, r *http.Request, service, operation string) *logrus.Entry {
	return logger.Logger(ctx).
		WithFields(logrus.Fields{
			"Service":   service,
			"Operation": operation,
			"UserAgent": r.UserAgent(),
			"IP":        r.RemoteAddr,
			"URI":       r.RequestURI,
		})

}
