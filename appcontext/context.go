// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package appcontext

import (
	"context"
	"io"
	"os"

	"github.com/condensat/bank-core"

	"github.com/condensat/bank-core/security"
	"github.com/condensat/bank-core/security/utils"

	log "github.com/sirupsen/logrus"
)

const (
	AppNameKey = iota
	LoggerKey
	ProcessusGrabberKey
	SecureIDKey
	WriterKey
	LogLevelKey
	CacheKey
	MessagingKey
	DatabaseKey

	privateKeySaltKey = security.KeyPrivateKeySalt
	hasherWorkerKey   = security.KeyHasherWorker
)

// WithAppName returns a context with the Application name set
func WithAppName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, AppNameKey, name)
}

// WithLogger returns a context with the log Writer set
func WithLogger(ctx context.Context, database bank.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, database)
}

// WithWriter returns a context with the log Writer set
func WithWriter(ctx context.Context, writer io.Writer) context.Context {
	return context.WithValue(ctx, WriterKey, writer)
}

// WithLogLevel returns a context with the LogLevel set
func WithLogLevel(ctx context.Context, level string) context.Context {
	return context.WithValue(ctx, LogLevelKey, level)
}

// WithMessaging returns a context with the messaging set
func WithCache(ctx context.Context, cache bank.Cache) context.Context {
	return context.WithValue(ctx, CacheKey, cache)
}

// WithMessaging returns a context with the messaging set
func WithMessaging(ctx context.Context, messaging bank.Messaging) context.Context {
	return context.WithValue(ctx, MessagingKey, messaging)
}

// WithDatabase returns a context with the database set
func WithDatabase(ctx context.Context, db bank.Database) context.Context {
	return context.WithValue(ctx, DatabaseKey, db)
}

// WithHasherWorker returns a context with the password worker set
func WithHasherWorker(ctx context.Context, options HasherOptions) context.Context {
	worker := security.NewHasherWorker(ctx, options.Time, options.Memory, options.Thread)
	go worker.Run(ctx, options.NumWorker)
	return context.WithValue(ctx, hasherWorkerKey, worker)
}

func WithOptions(ctx context.Context, options Options) context.Context {
	ctx = WithAppName(ctx, options.AppName)
	ctx = WithLogLevel(ctx, options.LogLevel)

	// generate random seed to hash private key and seed at runtime
	ctx = context.WithValue(ctx, privateKeySaltKey, utils.GenerateRandN(32))

	// Store PasswordHashSeed in context
	if len(options.PasswordHashSeed) == 0 {
		options.PasswordHashSeed = getEnv("PasswordHashSeed", "")
	}

	ctx = security.PasswordHashSeedContext(ctx, SecretOrPassword(options.PasswordHashSeed))
	os.Unsetenv("PasswordHashSeed")
	options.PasswordHashSeed = ""

	return ctx
}

func WithProcessusGrabber(ctx context.Context, grabber bank.Worker) context.Context {
	go grabber.Run(ctx, 1)
	return context.WithValue(ctx, ProcessusGrabberKey, grabber)
}

func WithSecureID(ctx context.Context, secureID bank.SecureID) context.Context {
	return context.WithValue(ctx, SecureIDKey, secureID)
}

func AppName(ctx context.Context) string {
	if ctxAppName, ok := ctx.Value(AppNameKey).(string); ok {
		return ctxAppName
	}
	return "NoAppName"
}

func Level(ctx context.Context) log.Level {
	if ctxLogLevel, ok := ctx.Value(LogLevelKey).(string); ok {
		if level, err := log.ParseLevel(ctxLogLevel); err == nil {
			return level
		}
	}
	return log.WarnLevel
}

func Logger(ctx context.Context) bank.Logger {
	if ctxLogger, ok := ctx.Value(LoggerKey).(bank.Logger); ok {
		return ctxLogger
	}
	return nil
}

func Cache(ctx context.Context) bank.Cache {
	if ctxCache, ok := ctx.Value(CacheKey).(bank.Cache); ok {
		return ctxCache
	}
	return nil
}

func Messaging(ctx context.Context) bank.Messaging {
	if ctxMessaging, ok := ctx.Value(MessagingKey).(bank.Messaging); ok {
		return ctxMessaging
	}
	return nil
}

func Database(ctx context.Context) bank.Database {
	if ctxDatabase, ok := ctx.Value(DatabaseKey).(bank.Database); ok {
		return ctxDatabase
	}
	return nil
}

func HasherWorker(ctx context.Context) bank.Worker {
	if ctxWorker, ok := ctx.Value(hasherWorkerKey).(bank.Worker); ok {
		return ctxWorker
	}
	return nil
}

func SecureID(ctx context.Context) bank.SecureID {
	if ctxSecureID, ok := ctx.Value(SecureIDKey).(bank.SecureID); ok {
		return ctxSecureID
	}
	return nil
}
