// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Logger grabber fetch entries from redis
package main

import (
	"context"
	"flag"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
)

type Args struct {
	AppName  string
	LogLevel string
	Redis    logger.RedisOptions

	WithDatabase bool
	Database     logger.DatabaseOptions
}

func parseArgs() Args {
	var args Args
	flag.StringVar(&args.AppName, "appName", "LogGrabber", "Application Name")
	flag.StringVar(&args.LogLevel, "log", "warning", "Log level [trace, debug, info, warning, error]")

	flag.StringVar(&args.Redis.HostName, "redisHost", "localhost", "Redis hostName (default 'localhost')")
	flag.IntVar(&args.Redis.Port, "redisPort", 6379, "Redis port (default 6379)")

	flag.BoolVar(&args.WithDatabase, "withDatabase", false, "Store log to database (default false)")
	flag.StringVar(&args.Database.HostName, "dbHost", "localhost", "Database hostName (default 'localhost')")
	flag.IntVar(&args.Database.Port, "dbPort", 3306, "Database port (default 3306)")
	flag.StringVar(&args.Database.User, "dbUser", "condensat", "Database user (condensat)")
	flag.StringVar(&args.Database.Password, "dbPassword", "condensat", "Database user (condensat)")
	flag.StringVar(&args.Database.Database, "dbName", "condensat", "Database name (condensat)")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := appcontext.WithAppName(context.Background(), args.AppName)
	ctx = appcontext.WithLogLevel(ctx, args.LogLevel)

	var databaseLogger *logger.DatabaseLogger
	if args.WithDatabase {
		databaseLogger = logger.NewDatabaseLogger(args.Database)
		defer databaseLogger.Close()

		ctx = appcontext.WithLogger(ctx, databaseLogger)
	}

	redisLogger := logger.NewRedisLogger(args.Redis)
	// Start the log grabber
	redisLogger.Grab(ctx)
}
