// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"testing"
	"time"
)

func TestInvalidSessionID(t *testing.T) {
	const empty = SessionID("")
	if cstInvalidSessionID != empty {
		t.Errorf("Wrong cstInvalidSessionID = %v, want %v", cstInvalidSessionID, empty)

	}
}

func TestNewSessionID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		unwant SessionID
	}{
		{"notInvalid", cstInvalidSessionID},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSessionID(); got == tt.unwant {
				t.Errorf("NewSessionID() = %v, unwant %v", got, tt.unwant)
			}
		})
	}
}

func TestSessionInfo_Expired(t *testing.T) {
	t.Parallel()

	var zero time.Time

	type fields struct {
		SessionID  SessionID
		Expiration time.Time
		delta      time.Duration
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, true},
		{"invalid", fields{cstInvalidSessionID, zero, 0}, true},

		{"now", fields{NewSessionID(), time.Now(), 0}, true},
		{"utc", fields{NewSessionID(), time.Now().UTC(), 0}, true},
		{"past", fields{NewSessionID(), zero, -time.Microsecond}, true},
		{"futur", fields{NewSessionID(), zero, 1 * time.Microsecond}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			expiration := tt.fields.Expiration
			if !expiration.After(zero) && tt.fields.delta > 0 {
				now := time.Now().UTC()
				expiration = now.Add(tt.fields.delta)
			}
			s := &SessionInfo{
				SessionID:  tt.fields.SessionID,
				Expiration: expiration,
			}
			if got := s.Expired(); got != tt.want {
				t.Errorf("SessionInfo.Expired() = %v, want %v", s.Expiration, time.Now())
			}
		})
	}
}

func TestSessionInfo_Encode(t *testing.T) {
	t.Parallel()

	var zero time.Time

	type fields struct {
		SessionID  SessionID
		Expiration time.Time
		delta      time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{"default", fields{}, 103, false},
		{"invalid", fields{cstInvalidSessionID, time.Time{}, 0}, 103, false},

		{"now", fields{NewSessionID(), time.Now(), 0}, 158, false},
		{"utc", fields{NewSessionID(), time.Now().UTC(), 0}, 158, false},
		{"past", fields{NewSessionID(), zero, -time.Microsecond}, 141, false},
		{"futur", fields{NewSessionID(), zero, 1 * time.Microsecond}, 158, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			expiration := tt.fields.Expiration
			if !expiration.After(zero) && tt.fields.delta > 0 {
				now := time.Now().UTC()
				expiration = now.Add(tt.fields.delta)
			}
			s := &SessionInfo{
				SessionID:  tt.fields.SessionID,
				Expiration: expiration,
			}
			got, err := s.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("SessionInfo.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("SessionInfo.Encode() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestSessionInfo_Decode(t *testing.T) {
	t.Parallel()

	var s0 SessionInfo
	d0, _ := s0.Encode()

	s1 := SessionInfo{}
	d1, _ := s1.Encode()

	s2 := SessionInfo{42, cstTestRemoteAddrSample, NewSessionID(), time.Now()}
	d2, _ := s2.Encode()

	s3 := SessionInfo{42, cstTestRemoteAddrSample, NewSessionID(), time.Now().UTC()}
	d3, _ := s3.Encode()

	s4 := SessionInfo{42, cstTestRemoteAddrSample, NewSessionID(), time.Now().UTC().Add(-time.Microsecond)}
	d4, _ := s4.Encode()

	s5 := SessionInfo{42, cstTestRemoteAddrSample, NewSessionID(), time.Now().UTC().Add(time.Microsecond)}
	d5, _ := s5.Encode()

	var zero [0]byte
	var empty [32]byte

	type fields struct {
		SessionID  SessionID
		Expiration time.Time
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    SessionInfo
		wantErr bool
	}{
		{"default", fields{}, args{d0}, s0, false},
		{"invalid", fields{}, args{d1}, s1, false},

		{"nil", fields{}, args{nil}, s0, true},
		{"zero", fields{}, args{zero[:]}, s0, true},
		{"empty", fields{}, args{empty[:]}, s0, true},

		{"now", fields{}, args{d2}, s2, false},
		{"utc", fields{}, args{d3}, s3, false},
		{"past", fields{}, args{d4}, s4, false},
		{"futur", fields{}, args{d5}, s5, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			s := &SessionInfo{
				SessionID:  tt.fields.SessionID,
				Expiration: tt.fields.Expiration,
			}
			if err := s.Decode(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SessionInfo.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if s.SessionID != tt.want.SessionID {
				t.Errorf("Decode() = %v, want %v", s.SessionID, tt.want.SessionID)
			}

			if !s.Expiration.Equal(tt.want.Expiration) {
				t.Errorf("Decode() = %v, want %v", s.Expiration, tt.want.Expiration)
			}
		})
	}
}
