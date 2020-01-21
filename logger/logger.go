// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"context"
	"io"

	"github.com/condensat/bank-core/appcontext"

	"github.com/sirupsen/logrus"
)

const (
	appKey = iota
	databaseKey
	writerKey
	logLevelKey
	messagingKey
)

func Logger(ctx context.Context) *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(appcontext.Level(ctx))
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logrus.NewEntry(logger)

	if ctx == nil {
		return entry
	}

	if ctxApp, ok := ctx.Value(appKey).(string); ok {
		entry = entry.WithField("app", ctxApp)
	}
	if ctxWriter, ok := ctx.Value(writerKey).(io.Writer); ok {
		logger.SetOutput(ctxWriter)
	}
	if ctxLogLevel, ok := ctx.Value(logLevelKey).(string); ok {
		if level, err := logrus.ParseLevel(ctxLogLevel); err == nil {
			logger.SetLevel(level)
		}
	}

	return entry
}
