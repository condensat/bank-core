// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/monitor"
	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/networking/ratelimiter"

	"github.com/condensat/bank-core/cache"

	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/messaging/provider"
	mprovider "github.com/condensat/bank-core/messaging/provider"

	"github.com/condensat/bank-core/monitor/tasks"
)

type StackMonitor struct {
	Port int

	PeerRequestPerSecond ratelimiter.RateLimitInfo

	CorsAllowedDomain string
}

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  mprovider.NatsOptions

	StackMonitor StackMonitor
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "StackMonitor")

	cache.OptionArgs(&args.Redis)
	mprovider.OptionArgs(&args.Nats)

	flag.IntVar(&args.StackMonitor.Port, "port", 4000, "Stack monitor port (default 4000)")

	args.StackMonitor.PeerRequestPerSecond = networking.DefaultPeerRequestPerSecond
	flag.IntVar(&args.StackMonitor.PeerRequestPerSecond.Rate, "peerRateLimit", 500, "Rate limit rate, per second, per peer connection (default 500)")

	flag.StringVar(&args.StackMonitor.CorsAllowedDomain, "corsAllowedDomain", "condensat.space", "Cors Allowed Domain (default condensat.space)")

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

	ctx = networking.RegisterRateLimiter(ctx, args.StackMonitor.PeerRequestPerSecond)

	var stackMonitor tasks.StackMonitor
	stackMonitor.Run(ctx, args.StackMonitor.Port, corsAllowedOrigins(args.StackMonitor.CorsAllowedDomain))
}

func corsAllowedOrigins(corsAllowedDomain string) []string {
	// sub-domains wildcard
	return []string{fmt.Sprintf("https://%s.%s", "*", corsAllowedDomain)}
}
