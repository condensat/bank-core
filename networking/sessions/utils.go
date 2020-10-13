// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"time"
)

func makeTimestampMillis(ts time.Time) int64 {
	return ts.UnixNano() / int64(time.Millisecond)
}

func fromTimestampMillis(timestamp int64) time.Time {
	return time.Unix(0, int64(timestamp)*int64(time.Millisecond)).UTC()
}
