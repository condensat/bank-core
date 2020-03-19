// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package utils

import (
	"context"
	"time"
)

func Scheduler(ctx context.Context, tick time.Duration, shift time.Duration) <-chan time.Time {
	if tick < 100*time.Millisecond || shift < 0 {
		panic("Invalid Scheduler arguments")
	}

	afterNextTick := func(tick time.Duration, shift time.Duration) <-chan time.Time {
		now := time.Now()
		nextTick := now.Add(tick).Truncate(tick)
		return time.After(nextTick.Sub(now) + shift)
	}

	next := make(chan time.Time)
	go func() {
		for {
			select {
			case <-afterNextTick(tick, shift):
				next <- time.Now().UTC()

			case <-ctx.Done():
				close(next)
				return
			}
		}
	}()

	return next
}
