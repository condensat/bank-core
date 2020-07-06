// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/condensat/bank-core/appcontext"

	"github.com/bsm/redislock"
)

const (
	DefaultLockTTL = 30 * time.Second

	MinimumLockTTL = 100 * time.Millisecond
	MaximumLockTTL = 5 * time.Minute
)

var (
	ErrRedisMutexNotFound = errors.New("RedisMutex Not Found")
	ErrLockError          = errors.New("Failed to acquire lock")
)

type Lock interface {
	Unlock()
}

type Mutex interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (Lock, error)
}

type RedisMutex struct {
	locker *redislock.Client
}

func NewRedisMutex(ctx context.Context) Mutex {
	client := ToRedis(appcontext.Cache(ctx))
	if client == nil {
		panic("Invalid Redis client")
	}

	locker := redislock.New(client)
	if locker == nil {
		panic("Invalid Redislock client")
	}

	return &RedisMutex{
		locker: locker,
	}
}

type RedisLock struct {
	lock *redislock.Lock
}

func (p *RedisLock) Unlock() {
	if p.lock != nil {
		_ = p.lock.Release()
	}
}

func (p *RedisMutex) Lock(ctx context.Context, key string, ttl time.Duration) (Lock, error) {
	if p.locker == nil {
		return nil, errors.New("Invalid locker")
	}

	if ttl < MinimumLockTTL {
		ttl = MinimumLockTTL
	}
	if ttl > MaximumLockTTL {
		ttl = MaximumLockTTL
	}

	const keyPrefix = "accounting.RedisMutex"

	lock, err := p.locker.Obtain(fmt.Sprintf("%s.%s", keyPrefix, key), ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 100),
		Metadata:      keyPrefix,
	})

	return &RedisLock{
		lock: lock,
	}, err
}

// Helper functions

func lockKeyString(prefix string, value interface{}) string {
	if prefix == "" {
		prefix = "lock"
	}
	return fmt.Sprintf("%s.%v", prefix, value)
}

func lockKeyGeneric(name string) string {
	return lockKeyString("lock.Name", name)
}

func lockKeyUserID(userID uint64) string {
	return lockKeyString("lock.User", userID)
}

func lockKeyAccountID(accountID uint64) string {
	return lockKeyString("lock.Account", accountID)
}

func lockKeyChain(chain string) string {
	return lockKeyString("lock.Chain", chain)
}

func lockKeyBatchNetwork(batchNetwork string) string {
	return lockKeyString("lock.BatchNetwork", batchNetwork)
}

func LockGeneric(ctx context.Context, name string) (Lock, error) {
	mutex := RedisMutexFromContext(ctx)
	if mutex == nil {
		return nil, ErrRedisMutexNotFound
	}
	return mutex.Lock(ctx, lockKeyGeneric(name), DefaultLockTTL)
}

func LockUser(ctx context.Context, userID uint64) (Lock, error) {
	mutex := RedisMutexFromContext(ctx)
	if mutex == nil {
		return nil, ErrRedisMutexNotFound
	}
	return mutex.Lock(ctx, lockKeyUserID(userID), DefaultLockTTL)
}

func LockAccount(ctx context.Context, accountID uint64) (Lock, error) {
	mutex := RedisMutexFromContext(ctx)
	if mutex == nil {
		return nil, ErrRedisMutexNotFound
	}
	return mutex.Lock(ctx, lockKeyAccountID(accountID), DefaultLockTTL)
}

func LockChain(ctx context.Context, chain string) (Lock, error) {
	mutex := RedisMutexFromContext(ctx)
	if mutex == nil {
		return nil, ErrRedisMutexNotFound
	}
	return mutex.Lock(ctx, lockKeyChain(chain), DefaultLockTTL)
}

func LockBatchNetwork(ctx context.Context, batchNetwork string) (Lock, error) {
	mutex := RedisMutexFromContext(ctx)
	if mutex == nil {
		return nil, ErrRedisMutexNotFound
	}
	return mutex.Lock(ctx, lockKeyBatchNetwork(batchNetwork), DefaultLockTTL)
}
