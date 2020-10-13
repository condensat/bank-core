// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package networking

import (
	"context"
	"net/http"

	"github.com/condensat/bank-core/logger"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func AppendRequestLog(log *logrus.Entry, r *http.Request) *logrus.Entry {
	return log.WithFields(logrus.Fields{
		"UserAgent": r.UserAgent(),
		"IP":        RequesterIP(r),
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

func CreateCorsOptions(corsAllowedOrigins []string) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: corsAllowedOrigins,
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowedMethods: []string{
			http.MethodPost,
		},
		MaxAge:           1000,
		AllowCredentials: true,
	})

}
