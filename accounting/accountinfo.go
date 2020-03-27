// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package accounting

import (
	"context"
	"time"
)

func ListUserAccounts(ctx context.Context, userID uint64) ([]AccountInfo, error) {
	var result []AccountInfo
	return result, nil
}

func GetAccountHistory(ctx context.Context, accountID uint64, from, to time.Time) ([]AccountEntry, error) {
	var result []AccountEntry
	return result, nil
}
