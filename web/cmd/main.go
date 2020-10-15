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
	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/web"

	"github.com/condensat/bank-core/cache"

	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/messaging/provider"
	mprovider "github.com/condensat/bank-core/messaging/provider"

	"github.com/condensat/bank-core/networking/ratelimiter"
)

type WebApp struct {
	Port                  int
	Directory             string
	SinglePageApplication bool

	PeerRequestPerSecond ratelimiter.RateLimitInfo
	OpenSessionPerMinute ratelimiter.RateLimitInfo
}

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  mprovider.NatsOptions

	WebApp WebApp
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankWebApp")

	cache.OptionArgs(&args.Redis)
	mprovider.OptionArgs(&args.Nats)

	flag.IntVar(&args.WebApp.Port, "port", 4420, "BankWebApp http port (default 4420)")
	flag.StringVar(&args.WebApp.Directory, "webDirectory", "/var/www", "BankWebApp http web directory (default /var/www)")
	flag.BoolVar(&args.WebApp.SinglePageApplication, "spa", false, "Is Single Page Application (default false")

	args.WebApp.PeerRequestPerSecond = networking.DefaultPeerRequestPerSecond
	flag.IntVar(&args.WebApp.PeerRequestPerSecond.Rate, "peerRateLimit", 20, "Rate limit rate, per second, per peer connection (default 20)")

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

	ctx = networking.RegisterRateLimiter(ctx, args.WebApp.PeerRequestPerSecond)

	var web web.Web
	web.Run(ctx, args.WebApp.Port, args.WebApp.Directory, args.WebApp.SinglePageApplication)
}
