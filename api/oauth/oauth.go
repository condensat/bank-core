// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package oauth

import (
	"errors"

	"github.com/joho/godotenv"
)

var (
	ErrInvalidOAuthKeys = errors.New("Invalid OAuth keys file")
)

type Options struct {
	Keys string
}

func Init(options Options) error {
	err := godotenv.Overload(options.Keys)
	if err != nil {
		return ErrInvalidOAuthKeys
	}
	return nil
}
