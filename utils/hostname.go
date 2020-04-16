// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package utils

import (
	"os"
)

func Hostname() string {
	var err error
	host, err := os.Hostname()
	if err != nil {
		host = "Unknown"
	}
	return host
}
