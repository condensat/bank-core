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

func OptionArgs(args *Options) {
	if args == nil {
		panic("Invalid database args")
	}

	flag.StringVar(&args.HostName, "dbHost", "db", "Database hostName (default 'db')")
	flag.IntVar(&args.Port, "dbPort", 3306, "Database port (default 3306)")
	flag.StringVar(&args.User, "dbUser", "condensat", "Database user (condensat)")
	flag.StringVar(&args.Password, "dbPassword", "condensat", "Database user (condensat)")
	flag.StringVar(&args.Database, "dbName", "condensat", "Database name (condensat)")
	flag.BoolVar(&args.EnableLogging, "enableLogging", false, "Enable database logging (false")
}
