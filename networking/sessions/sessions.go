// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const (
	KeySessions = "Api.Sessions"
)

type Session struct {
	rdb *redis.Client
}

func NewSession(ctx context.Context) *Session {
	rdb := cache.ToRedis(cache.FromContext(ctx))
	return &Session{
		rdb: rdb,
	}
}

func (s *Session) Count(ctx context.Context) (int, error) {
	log := logger.Logger(ctx).WithField("Method", "sessions.Session.Count")
	rdb := s.rdb

	keysWildcard := sessionKey("api", "sessions", "*")

	// Todo: optimize session count
	keys, err := rdb.Keys(ctx, keysWildcard).Result()
	if err != nil {
		log.WithError(err).
			WithField("Wildcard", keysWildcard).
			Error("Key count failed")
		return -1, ErrCache
	}
	return len(keys), nil
}

func (s *Session) CreateSession(ctx context.Context, userID uint64, remoteAddr string, duration time.Duration) (SessionID, error) {
	log := logger.Logger(ctx).WithField("Method", "sessions.Session.CreateSession")
	rdb := s.rdb

	log = log.WithFields(logrus.Fields{
		"UserID":     userID,
		"RemoteAddr": remoteAddr,
		"Duration":   duration,
	})

	if userID == cstInvalidUserID {
		log.Trace("Invalid userID")
		return cstInvalidSessionID, ErrInvalidUserID
	}

	if remoteAddr == cstInvalidRemoteAddr {
		log.Trace("Invalid remoteAddr")
		return cstInvalidSessionID, ErrInvalidUserID
	}

	if duration < time.Second {
		log.Trace("Invalid duration")
		return cstInvalidSessionID, ErrInvalidDuration
	}

	sessionID := NewSessionID()
	log = log.WithField("SessionID", sessionID)

	si, err := pushSession(ctx, rdb, userID, remoteAddr, sessionID, duration)
	if err != nil {
		log.Trace("Failed to push session to cache")
		return cstInvalidSessionID, err
	}

	log.WithField("Expiration", si.Expiration).
		Trace("New session created")

	return sessionID, nil
}

func (s *Session) sessionInfo(ctx context.Context, sessionID SessionID) SessionInfo {
	log := logger.Logger(ctx).WithField("Method", "sessions.Session.sessionInfo")
	rdb := s.rdb

	if sessionID == cstInvalidSessionID {
		return SessionInfo{}
	}
	log = log.WithField("SessionID", sessionID)

	// get session from cache
	si, err := fetchSession(ctx, rdb, sessionID)
	if err != nil {
		log.WithError(err).
			Trace("fetchSession failed")
		return SessionInfo{}
	}

	return si
}

func (s *Session) IsSessionValid(ctx context.Context, sessionID SessionID) bool {
	si := s.sessionInfo(ctx, sessionID)

	return !si.Expired()
}

func (s *Session) UserSession(ctx context.Context, sessionID SessionID) uint64 {
	log := logger.Logger(ctx).WithField("Method", "sessions.Session.UserSession")
	rdb := s.rdb

	if sessionID == cstInvalidSessionID {
		return cstInvalidUserID
	}
	log = log.WithField("SessionID", sessionID)

	// get session from cache
	si, err := fetchSession(ctx, rdb, sessionID)
	if err != nil {
		log.WithError(err).
			Trace("fetchSession failed")
		return cstInvalidUserID
	}

	return si.UserID
}

func IsSessionValid(sessionID SessionID) bool {
	return sessionID != cstInvalidSessionID
}

func IsUserValid(userID uint64) bool {
	return userID != cstInvalidUserID
}

func (s *Session) ExtendSession(ctx context.Context, remoteAddr string, sessionID SessionID, duration time.Duration) (uint64, error) {
	log := logger.Logger(ctx).WithField("Method", "sessions.Session.ExtendSession")
	rdb := s.rdb

	if remoteAddr == cstInvalidRemoteAddr {
		return cstInvalidUserID, ErrInvalidRemoteAddr
	}
	if sessionID == cstInvalidSessionID {
		return cstInvalidUserID, ErrInvalidSessionID
	}
	log = log.WithField("SessionID", sessionID)

	if duration <= 0 {
		log.WithField("Duration", duration).
			Trace("Invalid duration")
		return cstInvalidUserID, ErrInvalidDuration
	}

	// get session from cache
	si, err := fetchSession(ctx, rdb, sessionID)
	if err != nil {
		log.WithError(err).
			Trace("fetchSession failed")
		return cstInvalidUserID, ErrInvalidSessionID
	}
	log = log.WithFields(logrus.Fields{
		"UserID":     si.UserID,
		"RemoteAddr": si.RemoteAddr,
		"Duration":   duration,
	})

	// check for IP change
	if si.RemoteAddr != remoteAddr {
		log.WithField("NewRemoteAddr", remoteAddr).
			Trace("RemoteAddr has changed")
		// return userID with error
		return si.UserID, ErrRemoteAddrChanged
	}
	// do not renew expired session
	if si.Expired() {
		log.WithField("Expiration", si.Expiration).
			Trace("Session is expired")
		return cstInvalidUserID, ErrSessionExpired
	}

	si, err = pushSession(ctx, rdb, si.UserID, si.RemoteAddr, sessionID, duration)
	if err != nil {
		log.WithError(err).
			Trace("Failed to push session to cache")
		return cstInvalidUserID, err
	}

	log.WithField("Expiration", si.Expiration).
		Trace("Session extended")

	return si.UserID, nil
}

func (s *Session) InvalidateSession(ctx context.Context, sessionID SessionID) error {
	log := logger.Logger(ctx).WithField("Method", "sessions.Session.ExtendSession")
	rdb := s.rdb

	if sessionID == cstInvalidSessionID {
		return ErrInvalidSessionID
	}
	log = log.WithField("SessionID", sessionID)

	// get session from cache
	si, err := fetchSession(ctx, rdb, sessionID)
	if err != nil {
		log.WithError(err).
			Trace("fetchSession failed")
		return ErrInvalidSessionID
	}
	log = log.WithField("UserID", si.UserID)

	if si.Expired() {
		// NOOP
		return nil
	}

	duration := time.Duration(0)
	si, err = pushSession(ctx, rdb, cstInvalidUserID, cstInvalidRemoteAddr, sessionID, duration)
	if err != nil {
		log.WithError(err).
			WithField("Duration", duration).
			Trace("Failed to push session to cache")
		return err
	}

	log.WithField("Expiration", si.Expiration).
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

func pushSession(ctx context.Context, rdb *redis.Client, userID uint64, remoteAddr string, sessionID SessionID, duration time.Duration) (SessionInfo, error) {
	now := time.Now().UTC()
	expired := now.Add(-time.Nanosecond)
	si := SessionInfo{SessionID: cstInvalidSessionID, Expiration: expired}

	if sessionID == cstInvalidSessionID {
		return si, ErrInvalidSessionID
	}

	// expire session for invalid userID
	if userID == cstInvalidUserID {
		duration = 0
	}

	// expire session for invalid remoteAddr
	if remoteAddr == cstInvalidRemoteAddr {
		duration = 0
	}

	// update SessionInfo
	si.UserID = userID
	si.RemoteAddr = remoteAddr
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
		_, err := rdb.Del(ctx, key).Result()
		if err != nil {
			si.SessionID = cstInvalidSessionID
			si.Expiration = expired
			return si, ErrCache
		}
		return si, nil
	}

	// add 500ms to expiration key
	err = rdb.Set(ctx, key, data, duration+500*time.Millisecond).Err()
	if err != nil {
		si.SessionID = cstInvalidSessionID
		si.Expiration = expired
		return si, ErrCache
	}

	return si, nil
}

func fetchSession(ctx context.Context, rdb *redis.Client, sessionID SessionID) (SessionInfo, error) {
	if sessionID == cstInvalidSessionID {
		return SessionInfo{SessionID: cstInvalidSessionID}, ErrInvalidSessionID
	}
	key := sessionKey("api", "sessions", sessionID)
	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return SessionInfo{SessionID: cstInvalidSessionID}, ErrCache
	}
	var si SessionInfo
	now := time.Now().UTC()
	expired := now.Add(-time.Nanosecond)
	err = si.Decode(data)
	if err != nil {
		return SessionInfo{SessionID: cstInvalidSessionID, Expiration: expired}, ErrDecode
	}

	if si.SessionID != sessionID {
		return SessionInfo{SessionID: cstInvalidSessionID, Expiration: expired}, ErrInvalidSessionID
	}

	return si, nil
}
