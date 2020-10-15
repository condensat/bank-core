// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/monitor"

	"github.com/condensat/bank-core/cache"

	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/messaging/provider"
	mprovider "github.com/condensat/bank-core/messaging/provider"

	"github.com/condensat/bank-core/swap/liquid"
)

type Swap struct {
	ElementsConf string
}

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  mprovider.NatsOptions

	Swap Swap
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "LiquidSwap")

	cache.OptionArgs(&args.Redis)
	mprovider.OptionArgs(&args.Nats)

	flag.StringVar(&args.Swap.ElementsConf, "elementsConf", "/etc/liquidswap/elements.conf", "Elements conf file for RPC")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = cache.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = messaging.WithMessaging(ctx, provider.NewNats(ctx, args.Nats))
	ctx = appcontext.WithProcessusGrabber(ctx, monitor.NewProcessusGrabber(ctx, 15*time.Second))

	var swap liquid.Swap
	swap.Run(ctx, args.Swap.ElementsConf)
}
