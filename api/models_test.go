// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package api

import (
	"reflect"
	"testing"
)

func TestModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want int
	}{
		{"default", 9},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			if got := Models(); !reflect.DeepEqual(len(got), tt.want) {
				t.Errorf("Models() = %v, want %v", len(got), tt.want)
			}
		})
	}
}
