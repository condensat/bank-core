// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bank

import (
	"context"

	"github.com/condensat/bank-core/security/secureid"
)

type ServerOptions struct {
	Protocol string
	HostName string
	Port     int
}

type Worker interface {
	Run(ctx context.Context, numWorkers int)
}

type SecureID interface {
	ToSecureID(context string, value secureid.Value) (secureid.SecureID, error)
	FromSecureID(context string, secureID secureid.SecureID) (secureid.Value, error)

	ToString(secureID secureid.SecureID) string
	Parse(secureID string) secureid.SecureID
}
