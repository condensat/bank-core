// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package appcontext

import (
	"context"
	"io"

	"github.com/condensat/bank-core"

	log "github.com/sirupsen/logrus"
)

const (
	appKey = iota
	databaseKey
	writerKey
	logLevelKey
	messagingKey
)

// WithAppName returns a context with the Application name set
func WithAppName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, appKey, name)
}

// WithLogger returns a context with the log Writer set
func WithLogger(ctx context.Context, database bank.Logger) context.Context {
	return context.WithValue(ctx, databaseKey, database)
}

// WithWriter returns a context with the log Writer set
func WithWriter(ctx context.Context, writer io.Writer) context.Context {
	return context.WithValue(ctx, writerKey, writer)
}

// WithLogLevel returns a context with the LogLevel set
func WithLogLevel(ctx context.Context, level string) context.Context {
	return context.WithValue(ctx, logLevelKey, level)
}

// WithMessaging returns a context with the messaging set
func WithMessaging(ctx context.Context, messaging bank.Messaging) context.Context {
	return context.WithValue(ctx, messagingKey, messaging)
}

func Level(ctx context.Context) log.Level {
	if ctxLogLevel, ok := ctx.Value(logLevelKey).(string); ok {
		if level, err := log.ParseLevel(ctxLogLevel); err == nil {
			return level
		}
	}
	return log.WarnLevel
}

func Logger(ctx context.Context) bank.Logger {
	if ctxDatabase, ok := ctx.Value(databaseKey).(bank.Logger); ok {
		return ctxDatabase
	}
	return nil
}

func Messaging(ctx context.Context) bank.Messaging {
	if ctxMessaging, ok := ctx.Value(messagingKey).(bank.Messaging); ok {
		return ctxMessaging
	}
	return nil
}
