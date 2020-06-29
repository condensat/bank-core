// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"
	"time"
)

func TestBatch_IsComplete(t *testing.T) {
	type fields struct {
		ExecuteAfter time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"default", fields{}, true},

		{"not_complete", fields{time.Now().Add(time.Minute)}, false},
		{"now", fields{time.Now()}, true},
		{"complete", fields{time.Now().Add(-time.Minute)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Batch{
				ExecuteAfter: tt.fields.ExecuteAfter,
			}
			if got := p.IsComplete(); got != tt.want {
				t.Errorf("Batch.IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}
