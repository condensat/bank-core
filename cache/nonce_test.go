// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"testing"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
)

func TestResetNonce(t *testing.T) {
	ctx := context.Background()
	ctx = appcontext.WithCache(ctx, NewRedis(ctx, RedisOptions{
		ServerOptions: bank.ServerOptions{
			HostName: "redis",
			Port:     6379,
		},
	}))

	_, _ = Nonce(ctx, "ResetNonce.test", 42)

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"default", args{}, true},
		{"foo", args{"ResetNonce.foo"}, false},

		{"test", args{"ResetNonce.test"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResetNonce(ctx, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ResetNonce() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				nonce, _ := Nonce(ctx, tt.args.name, 0)
				if nonce != 0 {
					t.Errorf("ResetNonce() nonce still present = %v", nonce)
				}
			}
		})
	}
}

func TestNonce(t *testing.T) {
	ctx := context.Background()
	ctx = appcontext.WithCache(ctx, NewRedis(ctx, RedisOptions{
		ServerOptions: bank.ServerOptions{
			HostName: "redis",
			Port:     6379,
		},
	}))

	_ = ResetNonce(ctx, "Nonce.foo")
	_ = ResetNonce(ctx, "Nonce.test")
	_, _ = Nonce(ctx, "Nonce.test", 42)

	type args struct {
		name  string
		nonce uint64
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"default", args{}, 0, true},
		{"foo", args{"Nonce.foo", 42}, 0, false},

		{"test_get", args{"Nonce.test", 0}, 42, false},
		{"test_eq", args{"Nonce.test", 42}, 42, false},
		{"test_up", args{"Nonce.test", 43}, 42, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Nonce(ctx, tt.args.name, tt.args.nonce)
			if (err != nil) != tt.wantErr {
				t.Errorf("Nonce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Nonce() = %v, want %v", got, tt.want)
			}
		})
	}
}
