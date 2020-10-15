// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/condensat/bank-core/cache"

	"github.com/go-redis/redis/v8"
)

const cstTestRemoteAddrSample = "redis"

func TestNewSession(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *redis.Client
	}{
		{"new", args{ctx}, rdb},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSession(tt.args.ctx); !reflect.DeepEqual(got.rdb, tt.want) {
				t.Errorf("NewSession() = %v, want %v", got.rdb, tt.want)
			}
		})
	}
}

func TestSession_CreateSession(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	type fields struct {
		rdb *redis.Client
	}
	type args struct {
		ctx                     context.Context
		userID                  uint64
		cstTestRemoteAddrSample string
		duration                time.Duration
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      SessionID
		wantValid bool
		wantErr   bool
	}{
		{"default", fields{rdb}, args{ctx, 0, cstInvalidRemoteAddr, 0}, cstInvalidSessionID, false, true},

		// invalid UserID
		{"negative", fields{rdb}, args{ctx, 0, cstInvalidRemoteAddr, -time.Second}, cstInvalidSessionID, false, true},
		{"negative2", fields{rdb}, args{ctx, 0, cstInvalidRemoteAddr, -2 * time.Second}, cstInvalidSessionID, false, true},
		{"second", fields{rdb}, args{ctx, 0, cstInvalidRemoteAddr, time.Second}, cstInvalidSessionID, false, true},
		{"valid", fields{rdb}, args{ctx, 0, cstInvalidRemoteAddr, 2 * time.Second}, "non-empty-session", false, true},

		// with UserID
		{"negative", fields{rdb}, args{ctx, 42, cstTestRemoteAddrSample, -time.Second}, cstInvalidSessionID, false, true},
		{"negative2", fields{rdb}, args{ctx, 42, cstTestRemoteAddrSample, -2 * time.Second}, cstInvalidSessionID, false, true},
		{"second", fields{rdb}, args{ctx, 42, cstTestRemoteAddrSample, time.Second}, cstInvalidSessionID, true, false},
		{"valid", fields{rdb}, args{ctx, 42, cstTestRemoteAddrSample, 2 * time.Second}, "non-empty-session", true, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				rdb: tt.fields.rdb,
			}
			got, err := s.CreateSession(tt.args.ctx, tt.args.userID, tt.args.cstTestRemoteAddrSample, tt.args.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.CreateSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantValid != (got != cstInvalidSessionID) {
				t.Errorf("Session.CreateSession() = wrong SessionID %v, want %v", got != cstInvalidSessionID, tt.wantValid)
			}

			if s.IsSessionValid(ctx, got) != tt.wantValid {
				t.Errorf("Session.IsSessionValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestSession_IsSessionValid(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	s := NewSession(ctx)

	type fields struct {
		rdb *redis.Client
	}
	type args struct {
		ctx       context.Context
		sessionID SessionID
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		want               bool
		waitForExpire      time.Duration
		wantValidAfterWait bool
	}{
		{"default", fields{}, args{ctx, ""}, false, 0, false},
		{"invalid", fields{rdb}, args{ctx, cstInvalidSessionID}, false, 0, false},

		{"valid", fields{rdb}, args{ctx, createSession(ctx, s, time.Second)}, true, 0, true},

		{"not_expired", fields{rdb}, args{ctx, createSession(ctx, s, time.Second)}, true, 500 * time.Millisecond, true},
		{"expired", fields{rdb}, args{ctx, createSession(ctx, s, time.Second)}, true, time.Second, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				rdb: tt.fields.rdb,
			}
			if got := s.IsSessionValid(tt.args.ctx, tt.args.sessionID); got != tt.want {
				t.Errorf("Session.IsSessionValid() = %v, want %v", got, tt.want)
			}
			if tt.waitForExpire <= 0 {
				return
			}

			time.Sleep(tt.waitForExpire)

			if got := s.IsSessionValid(tt.args.ctx, tt.args.sessionID); got != tt.wantValidAfterWait {
				t.Errorf("Session.IsSessionValid() = %v, want %v", got, tt.wantValidAfterWait)
			}
		})
	}
}

func TestSession_UserSession(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	s := NewSession(ctx)

	type fields struct {
		rdb *redis.Client
	}
	type args struct {
		ctx       context.Context
		sessionID SessionID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{"default", fields{}, args{ctx, ""}, 0},
		{"invalid", fields{rdb}, args{ctx, cstInvalidSessionID}, 0},

		{"valid", fields{rdb}, args{ctx, createSession(ctx, s, time.Second)}, 42},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				rdb: tt.fields.rdb,
			}
			if got := s.UserSession(tt.args.ctx, tt.args.sessionID); got != tt.want {
				t.Errorf("Session.UserSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_ExtendSession(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	type fields struct {
		rdb *redis.Client
	}
	type args struct {
		ctx                     context.Context
		cstTestRemoteAddrSample string
		duration                time.Duration
		extend                  time.Duration
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		wantUserID    uint64
		waitForExpire time.Duration
	}{
		{"default", fields{}, args{ctx, cstInvalidRemoteAddr, 0, 0}, true, 0, 0},

		{"ip_changed", fields{rdb}, args{ctx, "10.0.0.1", time.Second, time.Second}, true, 42, 0},
		{"valid", fields{rdb}, args{ctx, cstTestRemoteAddrSample, time.Second, time.Second}, false, 42, 0},

		{"negative", fields{rdb}, args{ctx, cstTestRemoteAddrSample, time.Second, -time.Second}, true, 0, 0},

		{"not_expired", fields{rdb}, args{ctx, cstTestRemoteAddrSample, time.Second, time.Second}, false, 42, 500 * time.Millisecond},
		{"not_expired2", fields{rdb}, args{ctx, cstTestRemoteAddrSample, time.Second, time.Second}, false, 42, 900 * time.Millisecond},
		{"expired", fields{rdb}, args{ctx, cstTestRemoteAddrSample, time.Second, time.Second}, true, 0, 1100 * time.Millisecond},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				rdb: tt.fields.rdb,
			}

			sessionID := createSession(ctx, s, tt.args.duration)

			if tt.waitForExpire > 0 {
				time.Sleep(tt.waitForExpire)
			}

			userID, err := s.ExtendSession(tt.args.ctx, tt.args.cstTestRemoteAddrSample, sessionID, tt.args.extend)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.ExtendSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if userID != tt.wantUserID {
				t.Errorf("Session.ExtendSession() userID = %v, wantUserID %v", userID, tt.wantUserID)
			}
		})
	}
}

func TestSession_InvalidateSession(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	type fields struct {
		rdb *redis.Client
	}
	type args struct {
		ctx      context.Context
		duration time.Duration
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		waitForExpire time.Duration
	}{
		{"default", fields{rdb}, args{ctx, 0}, true, 0},
		{"negative", fields{rdb}, args{ctx, -time.Second}, true, 0},

		{"valid", fields{rdb}, args{ctx, time.Second}, false, 0},

		{"not_expired", fields{rdb}, args{ctx, time.Second}, false, 500 * time.Millisecond},
		{"not_expired2", fields{rdb}, args{ctx, time.Second}, false, 900 * time.Millisecond},

		// invalidate expired session must not return an error
		{"expired", fields{rdb}, args{ctx, time.Second}, false, 1100 * time.Millisecond},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				rdb: tt.fields.rdb,
			}

			sessionID := createSession(ctx, s, tt.args.duration)

			if sessionID != cstInvalidSessionID {
				if !s.IsSessionValid(ctx, sessionID) {
					t.Errorf("Session must be valid")
				}
			}

			if tt.waitForExpire > 0 {
				time.Sleep(tt.waitForExpire)
			}

			if err := s.InvalidateSession(tt.args.ctx, sessionID); (err != nil) != tt.wantErr {
				t.Errorf("Session.InvalidateSession() error = %v, wantErr %v", err, tt.wantErr)
			}

			if s.IsSessionValid(ctx, sessionID) {
				t.Errorf("Session.IsSessionValid() Invalidated session must not be valid")
			}
		})
	}
}

func Test_sessionKey(t *testing.T) {
	t.Parallel()

	type args struct {
		prefix    string
		key       string
		sessionID SessionID
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{}, ""},
		{"empty", args{"", "", ""}, ""},

		{"emptyPrefix", args{"", "key", "sessionID"}, ""},
		{"emptyKey", args{"prefix", "", "sessionID"}, ""},
		{"emptySession", args{"prefix", "key", ""}, ""},

		{"emptyPrefixKey", args{"", "", "sessionID"}, ""},
		{"emptyPrefixSession", args{"", "key", ""}, ""},
		{"KeySession", args{"prefix", "key", ""}, ""},

		{"valid", args{"prefix", "key", "sessionID"}, "prefix.key.sessionID"},

		{"pointBegin", args{".prefix", "key", "sessionID"}, "prefix.key.sessionID"},
		{"pointEnd", args{"prefix", "key", "sessionID."}, "prefix.key.sessionID"},

		{"slashBegin", args{"/prefix", "key", "sessionID"}, "prefix.key.sessionID"},
		{"slashEnd", args{"prefix", "key", "sessionID/"}, "prefix.key.sessionID"},

		{"spaces", args{"prefix ", "key ", "sessionID "}, "prefix.key.sessionID"},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := sessionKey(tt.args.prefix, tt.args.key, tt.args.sessionID); got != tt.want {
				t.Errorf("sessionKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pushSession(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	s1 := NewSessionID()
	s2 := NewSessionID()

	type args struct {
		rdb                     *redis.Client
		userID                  uint64
		cstTestRemoteAddrSample string
		sessionID               SessionID
		duration                time.Duration
	}
	tests := []struct {
		name        string
		args        args
		want        SessionID
		wantExpired bool
		wantErr     bool
	}{
		{"default", args{rdb, cstInvalidUserID, cstInvalidRemoteAddr, cstInvalidSessionID, 0}, cstInvalidSessionID, true, true},

		// Invalid UserID
		{"expired_user", args{rdb, cstInvalidUserID, cstInvalidRemoteAddr, s1, 0}, s1, true, false},
		{"sessionID_user", args{rdb, cstInvalidUserID, cstInvalidRemoteAddr, s2, time.Second}, s2, true, false},

		// Invalid RemoteAddr
		{"expired_addr", args{rdb, 42, cstInvalidRemoteAddr, s1, 0}, s1, true, false},
		{"sessionID_addr", args{rdb, 42, cstInvalidRemoteAddr, s2, time.Second}, s2, true, false},

		// Invalid UserID & RemoteAddr
		{"expired_user", args{rdb, cstInvalidUserID, cstTestRemoteAddrSample, s1, 0}, s1, true, false},
		{"sessionID_user", args{rdb, cstInvalidUserID, cstTestRemoteAddrSample, s2, time.Second}, s2, true, false},

		{"expired", args{rdb, 42, cstTestRemoteAddrSample, s1, 0}, s1, true, false},
		{"valid", args{rdb, 42, cstTestRemoteAddrSample, s2, time.Second}, s2, false, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := pushSession(ctx, tt.args.rdb, tt.args.userID, tt.args.cstTestRemoteAddrSample, tt.args.sessionID, tt.args.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("pushSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Expired() != tt.wantExpired {
				t.Errorf("pushSession() expired = %v, wantExpired %v", got.Expired(), tt.wantExpired)
				return
			}
			if !reflect.DeepEqual(got.SessionID, tt.want) {
				t.Errorf("pushSession() = %v, want %v", got, tt.want)
			}
			if tt.args.duration <= 0 {
				return
			}
		})
	}
}

func Test_fetchSession(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctx = cache.WithCache(ctx, createCache(ctx))
	rdb := cache.ToRedis(cache.FromContext(ctx))

	s := NewSession(ctx)

	s0 := createSession(ctx, s, 0)
	s1 := createSession(ctx, s, time.Second)

	type args struct {
		rdb       *redis.Client
		sessionID SessionID
	}
	tests := []struct {
		name    string
		args    args
		want    SessionID
		wantErr bool
	}{
		{"default", args{rdb, cstInvalidSessionID}, cstInvalidSessionID, true},
		{"expired", args{rdb, s0}, cstInvalidSessionID, true},
		{"valid", args{rdb, s1}, s1, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchSession(ctx, tt.args.rdb, tt.args.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.SessionID, tt.want) {
				t.Errorf("fetchSession() = %v, want %v", got.SessionID, tt.want)
			}
		})
	}
}

func createCache(ctx context.Context) cache.Cache {
	return cache.NewRedis(ctx, cache.RedisOptions{
		HostName: "redis",
		Port:     6379,
	})
}

func createSession(ctx context.Context, s *Session, d time.Duration) SessionID {
	sID, _ := s.CreateSession(ctx, 42, cstTestRemoteAddrSample, d)
	return sID
}
