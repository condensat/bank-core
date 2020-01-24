// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"github.com/condensat/bank-core"

	nats "github.com/nats-io/nats.go"
)

func ToNats(messaging bank.Messaging) *nats.Conn {
	nc := messaging.NC()
	return nc.(*nats.Conn)
}
