// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

import (
	"math"
	"syscall"
	"time"
)

type Clock struct {
	Start time.Time
	Clock time.Duration
}

func (p *Clock) Init() {
	p.Start = time.Now()
	p.Clock = clock()
}

func (p *Clock) CPU() float64 {
	clockSeconds := clock() - p.Clock
	realSeconds := time.Since(p.Start)

	ret := float64(clockSeconds) / float64(realSeconds) * 100.0
	return math.Round(ret*100.0) / 100.0
}

func clock() time.Duration {
	var ru syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	if err != nil {
		panic(err)
	}
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}
