// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"

	"github.com/condensat/bank-core/wallet/chain"
)

func UpdateRedisChain(ctx context.Context, chainsStates ...chain.ChainState) error {
	// Todo: store chains states into redis
	return nil
}
