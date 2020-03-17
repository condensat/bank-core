// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"
	"time"

	"github.com/condensat/bank-core/logger"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func makeTimestampMillis(ts time.Time) int64 {
	return ts.UnixNano() / int64(time.Millisecond)
}

func fromTimestampMillis(timestamp int64) time.Time {
	return time.Unix(0, int64(timestamp)*int64(time.Millisecond)).UTC()
}

func RequesterIP(r *http.Request) string {
	// Header added by reverse proxy
	const cfConnectingIP = "CF-Connecting-IP" // CloudFlare
	const xRealIP = "X-Real-Ip"               // traefik
	const xForwardedFor = "X-Forwarded-For"   // Generic proxy

	// Priority order
	if ips, ok := r.Header[cfConnectingIP]; ok && len(ips) > 0 {
		return ips[0]
	} else if ips, ok := r.Header[xRealIP]; ok && len(ips) > 0 {
		return ips[0]
	} else if ips, ok := r.Header[xForwardedFor]; ok && len(ips) > 0 {
		return ips[0]
	} else {
		return r.RemoteAddr // fallback with RemoteAddr
	}
}

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
