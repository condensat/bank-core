// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package utils

import (
	"fmt"
)

func EllipsisCentral(str string, limit int) string {
	if len(str) <= 2*limit {
		return str
	}
	return fmt.Sprintf("%s...%s", str[:limit], str[len(str)-limit:])
}
