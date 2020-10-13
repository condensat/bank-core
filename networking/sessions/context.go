// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"context"
	"errors"
)

var (
	ErrInternalError = errors.New("Internal Error")
)

func ContextSession(ctx context.Context) (*Session, error) {
	if ctxSession, ok := ctx.Value(KeySessions).(*Session); ok {
		return ctxSession, nil
	}
	return nil, ErrInternalError
}
