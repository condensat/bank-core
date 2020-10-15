// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package appcontext

import (
	"context"
)

type Worker interface {
	Run(ctx context.Context, numWorkers int)
}
