// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"flag"
)

type Options struct {
	HostName      string
	Port          int
	User          string
	Password      string
	Database      string
	EnableLogging bool
}

func DefaultOptions() Options {
	return Options{
		HostName:      "db",
		Port:          3306,
		User:          "condensat",
		Password:      "condensat",
		Database:      "condensat",
		EnableLogging: false,
	}
}

func OptionArgs(args *Options) {
	if args == nil {
		panic("Invalid database args")
	}

	defaults := DefaultOptions()
	flag.StringVar(&args.HostName, "dbHost", defaults.HostName, "Database hostName (default 'db')")
	flag.IntVar(&args.Port, "dbPort", defaults.Port, "Database port (default 3306)")
	flag.StringVar(&args.User, "dbUser", defaults.User, "Database user (condensat)")
	flag.StringVar(&args.Password, "dbPassword", defaults.Password, "Database user (condensat)")
	flag.StringVar(&args.Database, "dbName", defaults.Database, "Database name (condensat)")
	flag.BoolVar(&args.EnableLogging, "enableLogging", defaults.EnableLogging, "Enable database logging (false")
}
