// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/monitor"
)

func main() {
	var db database.Options
	database.OptionArgs(&db)
	flag.Parse()

	ctx := context.Background()
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(db))

	services, err := monitor.LastServicesStatus(ctx)
	if err != nil {
		panic(err)
	}
	for _, service := range services {
		fmt.Printf("  %+v\n", service)
	}
}
