// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
)

type Api int

func (p *Api) Run(ctx context.Context) {

	<-ctx.Done()
}
