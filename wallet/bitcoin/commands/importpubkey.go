// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
)

func ImportPubKey(ctx context.Context, rpcClient RpcClient, pubKey PubKey, label string, reindex bool) error {
	var noResult GenericJson
	err := callCommand(rpcClient, CmdImportPubKey, &noResult, pubKey, label, reindex)
	if err != nil {
		return err
	}

	return nil
}
