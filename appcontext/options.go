// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package appcontext

import (
	"flag"

	dotenv "github.com/joho/godotenv"
)

type HasherOptions struct {
	Time   int
	Memory int
	Thread int

	NumWorker int
}

type Options struct {
	AppName  string
	LogLevel string

	PasswordHashSeed string
	Hasher           HasherOptions
}

func OptionArgs(args *Options, defaultAppName string) {
	// import .env file to memory
	_ = dotenv.Load()

	if args == nil {
		panic("Invalid args options")
	}
	if len(defaultAppName) == 0 {
		defaultAppName = "NoAppName"
	}

	flag.StringVar(&args.AppName, "appName", defaultAppName, "Application Name")
	flag.StringVar(&args.LogLevel, "log", "warning", "Log level [trace, debug, info, warning, error]")

	flag.StringVar(&args.PasswordHashSeed, "hash_seed", "", "Seed used for hash salt")

	// Hasher parameters
	flag.IntVar(&args.Hasher.Time, "hasher_time", 3, "Hash iteration time (3)")
	flag.IntVar(&args.Hasher.Memory, "hasher_memory", 1<<16, "Hash allocated memory ( 16 MiB)")
	flag.IntVar(&args.Hasher.Thread, "hasher_thread", 4, "Hash threads (4)")

	flag.IntVar(&args.Hasher.NumWorker, "hasher_worker", 4, "Number of hasher workers")
}
