// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"
)

type SingleCall func(ctx context.Context) error

func InitSingleCall(ctx context.Context, name string) error {
	return ResetNonce(ctx, nonceName(name))
}

func ExecuteSingleCall(ctx context.Context, name string, call SingleCall) error {
	// get current nonce
	nonce, err := Nonce(ctx, nonceName(name), 0)
	if err != nil {
		return err
	}

	// critical section
	lock, err := LockGeneric(ctx, lockName(name))
	if err != nil {
		return err
	}
	defer lock.Unlock()

	// get last nonce within critical section
	lastNonce, err := Nonce(ctx, nonceName(name), nonce)
	if err != nil {
		return err
	}

	// skip call if nonce has changed
	if lastNonce != nonce {
		return nil
	}
	// increment nonce within critical section
	_, err = Nonce(ctx, nonceName(name), nonce+1)
	if err != nil {
		return err
	}

	// effective call
	return call(ctx)
}

func nonceName(name string) string {
	return name + ".nonce"
}

func lockName(name string) string {
	return name + ".lock"
}
