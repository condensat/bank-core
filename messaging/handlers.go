// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package messaging

import (
	"context"
	"errors"
)

var (
	ErrHandleRequest = errors.New("Request handle failed")
	ErrRequestFailed = errors.New("Request Failed")
)

type RequestHandler func(ctx context.Context, request BankObject) (BankObject, error)

func HandleRequest(ctx context.Context, appName string, message *Message, request BankObject, handle RequestHandler) (*Message, error) {
	err := FromMessage(message, request)
	if err != nil {
		return nil, err
	}

	resp, err := handle(ctx, request)
	if err != nil {
		return nil, err
	}

	message = ToMessage(appName, resp)
	if message == nil {
		err = ErrHandleRequest
	}

	return message, err
}

func RequestMessage(ctx context.Context, appName string, subject string, req, resp BankObject) error {
	messaging := FromContext(ctx)

	message := ToMessage(appName, req)

	message, err := messaging.Request(ctx, subject, message)
	if err != nil {
		return err
	}

	err = FromMessage(message, resp)
	if err != nil {
		return err
	}

	return nil
}
