// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"flag"

	"github.com/condensat/bank-core"
)

type NatsOptions struct {
	bank.ServerOptions
}

func OptionArgs(args *NatsOptions) {
	if args == nil {
		panic("Invalid args options")
	}

	flag.StringVar(&args.HostName, "natsHost", "localhost", "Nats hostName (default 'localhost')")
	flag.IntVar(&args.Port, "natsPort", 4222, "Nats port (default 4222)")
}
