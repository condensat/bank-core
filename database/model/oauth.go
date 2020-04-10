// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type OAuth struct {
	ID         ID     `gorm:"primary_key"`                                            // [PK] OAuth
	Provider   string `gorm:"unique_index:idx_prov_provid;not null;type:varchar(16)"` // [U] Provider name
	ProviderID string `gorm:"unique_index:idx_prov_provid;not null;type:varchar(64)"` // [U] Provider unique ID (UserID or NickName)
	UserID     UserID `gorm:"index;not null"`                                         // [FK] Reference to User table. Same user can have multiple providers
}

type OAuthData struct {
	OAuthID ID     `gorm:"unique_index;not null"`           // [FK] Reference to OAuth table
	Data    string `gorm:"type:json;not null;default:'{}'"` // goth User json data
}
