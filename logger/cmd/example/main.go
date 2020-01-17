// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// simply push log entry to redis
package main

import (
	"context"
	"flag"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"
)

type Args struct {
	AppName  string
	LogLevel string
	Redis    logger.RedisOptions
	Nats     messaging.NatsOptions
}

func parseArgs() Args {
	var args Args
	flag.StringVar(&args.AppName, "appName", "LoggerExample", "Application Name")
	flag.StringVar(&args.LogLevel, "log", "warning", "Log level [trace, debug, info, warning, error]")

	flag.StringVar(&args.Redis.HostName, "redisHost", "localhost", "Redis hostName (default 'localhost')")
	flag.IntVar(&args.Redis.Port, "redisPort", 6379, "Redis port (default 6379)")

	flag.StringVar(&args.Nats.HostName, "natsHost", "localhost", "Nats hostName (default 'localhost')")
	flag.IntVar(&args.Nats.Port, "natsPort", 4222, "Nats port (default 4222)")

	flag.Parse()

	return args
}

func echoHandler(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	logger.Logger(ctx).
		WithField("Subject", subject).
		WithField("Method", "echoHandler").
		Infof("-> %s", string(message.Data))

	return message, nil
}

func natsClient(ctx context.Context) {
	messaging := logger.ContextMessaging(ctx)
	messaging.SubscribeWorkers(ctx, "Example.Request", 8, echoHandler)

	log := logger.Logger(ctx)
	message := bank.NewMessage()
	message.Data = []byte("Hello, World!")

	for index := 0; index < 10; index++ {
		resp, err := messaging.Request(ctx, "Example.Request", message)
		if err != nil {
			log.
				WithError(err).
				Panicf("Request failed")
		}
		log.
			WithField("Method", "natsClient").
			Infof("<- %s", string(resp.Data))
	}
}

func main() {
	args := parseArgs()

	ctx := logger.WithAppName(context.Background(), args.AppName)
	ctx = logger.WithWriter(ctx, logger.NewRedisLogger(args.Redis))
	ctx = logger.WithLogLevel(ctx, args.LogLevel)
	ctx = logger.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))

	natsClient(ctx)
}
