// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package appcontext

import (
	"flag"
)

type Options struct {
	AppName  string
	LogLevel string
}

func OptionArgs(args *Options, defaultAppName string) {
	if args == nil {
		panic("Invalid args options")
	}
	if len(defaultAppName) == 0 {
		defaultAppName = "NoAppName"
	}

	flag.StringVar(&args.AppName, "appName", defaultAppName, "Application Name")
	flag.StringVar(&args.LogLevel, "log", "warning", "Log level [trace, debug, info, warning, error]")
}
