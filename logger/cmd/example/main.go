// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// simply push log entry to redis
package main

import (
	"context"
	"flag"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/cache"

	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/messaging/provider"
	mprovider "github.com/condensat/bank-core/messaging/provider"

	"github.com/condensat/bank-core/monitor"

	"github.com/sirupsen/logrus"
)

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  mprovider.NatsOptions
}

func parseArgs() Args {
	var args Args
	appcontext.OptionArgs(&args.App, "LoggerExample")

	cache.OptionArgs(&args.Redis)
	mprovider.OptionArgs(&args.Nats)

	flag.Parse()

	return args
}

func echoHandler(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "main.echoHandler")

	log.WithFields(logrus.Fields{
		"Subject": subject,
		"Method":  "echoHandler",
	}).Infof("-> %s", string(message.Data))

	return message, nil
}

func natsClient(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "main.natsClient")

	msging := messaging.FromContext(ctx)
	msging.SubscribeWorkers(ctx, "Example.Request", 8, echoHandler)

	message := messaging.NewMessage()
	message.Data = []byte("Hello, World!")

	for index := 0; index < 10; index++ {
		resp, err := msging.Request(ctx, "Example.Request", message)
		if err != nil {
			log.
				WithError(err).
				Panicf("Request failed")
		}
		log.Infof("<- %s", string(resp.Data))
	}
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = cache.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = messaging.WithMessaging(ctx, provider.NewNats(ctx, args.Nats))
	ctx = appcontext.WithProcessusGrabber(ctx, monitor.NewProcessusGrabber(ctx, 15*time.Second))

	natsClient(ctx)
}
