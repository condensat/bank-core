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
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/monitor"
	"github.com/condensat/bank-core/monitor/common"
)

func main() {
	var db database.Options
	database.OptionArgs(&db)
	flag.Parse()

	ctx := context.Background()
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(db))

	step := 15 * time.Second
	timeframe := 10 * time.Minute
	to := time.Now().UTC().Truncate(step)
	from := to.Add(-timeframe)
	round := time.Minute

	apps, err := monitor.ListServices(ctx, timeframe)
	if err != nil {
		panic(err)
	}

	var serviceMap = make(map[string][]common.ProcessInfo)
	for _, appName := range apps {
		services, err := monitor.LastServiceHistory(ctx, appName, from, to, step, round)
		if err != nil {
			panic(err)
		}

		for _, service := range services {
			serviceName := fmt.Sprintf("%s:%s", service.AppName, service.Hostname)
			serviceMap[serviceName] = append(serviceMap[serviceName], service)
		}
	}

	fmt.Printf("%d services:\n", len(serviceMap))
	for serviceName, history := range serviceMap {
		fmt.Printf("  %s: %d\n", serviceName, len(history))
		for _, info := range history {
			fmt.Printf("    %s, %5.2f %%, %5.1f MiB\n", info.Timestamp, info.CPUUsage, float64(info.MemAlloc)/float64(1<<20))
		}
	}
}
