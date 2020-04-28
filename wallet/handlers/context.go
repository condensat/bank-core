// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"errors"
)

const (
	ChainHandlerKey = "Key.ChainHandlerKey"
)

var (
	ErrInternalError = errors.New("Internal Error")
)

type ChainHandler interface {
	GetNewAddress(ctx context.Context, chain, account string) (string, error)
}

func ChainHandlerContext(ctx context.Context, chain ChainHandler) context.Context {
	return context.WithValue(ctx, ChainHandlerKey, chain)
}

func ChainHandlerFromContext(ctx context.Context) ChainHandler {
	if ctxChainHandler, ok := ctx.Value(ChainHandlerKey).(ChainHandler); ok {
		return ctxChainHandler
	}
	return nil
}
