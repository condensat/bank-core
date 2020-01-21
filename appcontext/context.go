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
	loggerKey
	writerKey
	logLevelKey
	messagingKey
	databaseKey
)

// WithAppName returns a context with the Application name set
func WithAppName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, appKey, name)
}

// WithLogger returns a context with the log Writer set
func WithLogger(ctx context.Context, database bank.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, database)
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

// WithDatabase returns a context with the database set
func WithDatabase(ctx context.Context, db bank.Database) context.Context {
	return context.WithValue(ctx, databaseKey, db)
}

func WithOptions(ctx context.Context, options Options) context.Context {
	ctx = WithAppName(ctx, options.AppName)
	ctx = WithLogLevel(ctx, options.LogLevel)
	return ctx
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
	if ctxDatabase, ok := ctx.Value(loggerKey).(bank.Logger); ok {
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

func Database(ctx context.Context) bank.Database {
	if ctxDatabase, ok := ctx.Value(databaseKey).(bank.Database); ok {
		return ctxDatabase
	}
	return nil
}
