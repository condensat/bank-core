// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func ImportBlindingKey(ctx context.Context, rpcClient RpcClient, address Address, blindingKey BlindingKey) error {
	var noResult GenericJson
	err := callCommand(rpcClient, CmdImportBlindingKey, &noResult, address, blindingKey)
	if err != nil {
		return err
	}

	return nil
}
