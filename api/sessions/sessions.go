// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"

	"github.com/go-redis/redis"
)

var (
	ErrInvalidDuration  = errors.New("Invalid Duration")
	ErrInvalidSessionID = errors.New("Invalid SessionID")
	ErrSessionExpired   = errors.New("Session Expired")
	ErrEncode           = errors.New("Encode Error")
	ErrDecode           = errors.New("Decode Error")
	ErrCache            = errors.New("Cache Error")
)

type Session struct {
	rdb *redis.Client
}

func NewSession(ctx context.Context) *Session {
	rdb := cache.ToRedis(appcontext.Cache(ctx))
	return &Session{
		rdb: rdb,
	}
}

func (s *Session) CreateSession(ctx context.Context, duration time.Duration) (SessionID, error) {
	rdb := s.rdb
	log := logger.Logger(ctx).WithField("Method", "api.Session.CreateSession")

	sessionID := NewSessionID()

	if duration < time.Second {
		log.
			WithField("SessionID", sessionID).
			WithField("Duration", duration).
			Debug("Invalid duration")
		return cstInvalidSessionID, ErrInvalidDuration
	}

	si, err := pushSession(rdb, sessionID, duration)
	if err != nil {
		log.
			WithError(err).
			WithField("SessionID", sessionID).
			WithField("Duration", duration).
			Debug("Failed to push session to cache")
		return cstInvalidSessionID, err
	}

	log.
		WithField("SessionID", si.SessionID).
		WithField("Expiration", si.Expiration).
		Trace("New session created")

	return sessionID, nil
}

func (s *Session) IsSessionValid(ctx context.Context, sessionID SessionID) bool {
	rdb := s.rdb
	log := logger.Logger(ctx).WithField("Method", "api.Session.IsSessionValid")
	if sessionID == cstInvalidSessionID {
		return false
	}

	// get session from cache
	si, err := fetchSession(rdb, sessionID)
	if err != nil {
		log.
			WithError(err).
			WithField("SessionID", sessionID).
			Error("Failed to get session from cache")
		return false
	}

	return !si.Expired()
}

func (s *Session) ExtendSession(ctx context.Context, sessionID SessionID, duration time.Duration) error {
	rdb := s.rdb
	log := logger.Logger(ctx).WithField("Method", "api.Session.ExtendSession")

	if duration <= 0 {
		log.
			WithField("SessionID", sessionID).
			WithField("Duration", duration).
			Debug("Invalid duration")
		return ErrInvalidDuration
	}

	// get session from cache
	si, err := fetchSession(rdb, sessionID)
	if err != nil {
		log.
			WithError(err).
			WithField("SessionID", sessionID).
			Error("Failed to get session from cache")
		return err
	}

	// do not renew expired session
	if si.Expired() {
		log.
			WithField("SessionID", si.SessionID).
			WithField("Expiration", si.Expiration).
			Debug("Session is expired")
		return ErrSessionExpired
	}

	si, err = pushSession(rdb, sessionID, duration)
	if err != nil {
		log.
			WithError(err).
			WithField("SessionID", si.SessionID).
			WithField("Duration", duration).
			Debug("Failed to push session to cache")
		return err
	}

	log.
		WithField("SessionID", si.SessionID).
		WithField("Expiration", si.Expiration).
		WithField("Duration", duration).
		Trace("Session extended")

	return nil
}

func (s *Session) InvalidateSession(ctx context.Context, sessionID SessionID) error {
	rdb := s.rdb
	log := logger.Logger(ctx).WithField("Method", "api.Session.InvalidateSession")
	if sessionID == cstInvalidSessionID {
		return ErrInvalidSessionID
	}

	// get session from cache
	si, err := fetchSession(rdb, sessionID)
	if err != nil {
		log.
			WithError(err).
			WithField("SessionID", sessionID).
			Error("Failed to get session from cache")
		return ErrCache
	}

	if si.Expired() {
		// NOOP
		return nil
	}

	duration := time.Duration(0)
	si, err = pushSession(rdb, sessionID, duration)
	if err != nil {
		log.
			WithError(err).
			WithField("SessionID", si.SessionID).
			WithField("Duration", duration).
			Debug("Failed to push session to cache")
		return err
	}

	log.
		WithField("SessionID", si.SessionID).
		WithField("Expiration", si.Expiration).
		WithField("Duration", duration).
		Trace("Session invalidated")

	return nil
}

func sessionKey(prefix, key string, sessionID SessionID) string {
	if len(prefix) == 0 || len(key) == 0 || len(sessionID) == 0 {
		return ""
	}
	str := fmt.Sprintf("%s.%s.%s", prefix, key, sessionID)
	str = strings.Trim(str, "./ ")
	str = strings.ReplaceAll(str, " ", "")
	return str

}

func pushSession(rdb *redis.Client, sessionID SessionID, duration time.Duration) (SessionInfo, error) {
	now := time.Now().UTC()
	expired := now.Add(-time.Nanosecond)
	si := SessionInfo{SessionID: cstInvalidSessionID, Expiration: expired}

	if sessionID == cstInvalidSessionID {
		return si, ErrInvalidSessionID
	}

	// update SessionInfo
	si.SessionID = sessionID
	si.Expiration = now.Add(duration)

	data, err := si.Encode()
	if err != nil {
		si.SessionID = cstInvalidSessionID
		si.Expiration = expired
		return si, ErrEncode
	}

	key := sessionKey("api", "sessions", si.SessionID)
	if si.Expired() {
		_, err := rdb.Del(key).Result()
		if err != nil {
			si.SessionID = cstInvalidSessionID
			si.Expiration = expired
			return si, ErrCache
		}
		return si, nil
	}

	// add 500ms to expiration key
	err = rdb.Set(key, data, duration+500*time.Millisecond).Err()
	if err != nil {
		si.SessionID = cstInvalidSessionID
		si.Expiration = expired
		return si, ErrCache
	}

	return si, nil
}

func fetchSession(rdb *redis.Client, sessionID SessionID) (SessionInfo, error) {
	if sessionID == cstInvalidSessionID {
		return SessionInfo{SessionID: cstInvalidSessionID}, ErrInvalidSessionID
	}
	key := sessionKey("api", "sessions", sessionID)
	data, err := rdb.Get(key).Bytes()
	if err != nil {
		return SessionInfo{SessionID: cstInvalidSessionID}, ErrCache
	}
	var si SessionInfo
	err = si.Decode(data)
	if err != nil {
		return SessionInfo{SessionID: cstInvalidSessionID}, ErrDecode
	}

	if si.SessionID != sessionID {
		return SessionInfo{SessionID: cstInvalidSessionID}, ErrInvalidSessionID
	}

	return si, nil
}
