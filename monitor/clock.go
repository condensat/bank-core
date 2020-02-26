// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package monitor

//#include <time.h>
import "C"

import (
	"math"
	"time"
)

type Clock struct {
	Start time.Time
	Ticks C.clock_t
}

func (p *Clock) Init() {
	p.Start = time.Now()
	p.Ticks = clock()
}

func (p *Clock) CPU() float64 {
	clockSeconds := float64(clock()-p.Ticks) / 1000000.0 // C.CLOCKS_PER_SEC == 1000000

	realSeconds := time.Since(p.Start).Seconds()
	ret := clockSeconds / realSeconds * 100.0
	return math.Round(ret*100.0) / 100.0
}

func clock() C.clock_t {
	return C.clock()
}
