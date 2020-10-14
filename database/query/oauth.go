// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package query

import (
	"errors"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidProvider   = errors.New("Invalid Provider")
	ErrInvalidProviderID = errors.New("Invalid ProviderID")
	ErrInvalidOAuthID    = errors.New("Invalid OAuthID")
	ErrInvalidOAuthData  = errors.New("Invalid OAuth Data")
)

// FindOrCreateOAuth
func FindOrCreateOAuth(db database.Context, oauth model.OAuth) (model.OAuth, error) {
	if len(oauth.Provider) == 0 {
		return model.OAuth{}, ErrInvalidProvider
	}
	if len(oauth.ProviderID) == 0 {
		return model.OAuth{}, ErrInvalidProviderID
	}
	if oauth.UserID == 0 {
		return model.OAuth{}, ErrInvalidUserID
	}

	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result model.OAuth
		err := gdb.
			Where(model.OAuth{
				Provider:   oauth.Provider,
				ProviderID: oauth.ProviderID,
			}).
			Assign(oauth).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.OAuth{}, database.ErrInvalidDatabase
	}
}

// CreateOrUpdateOAuthData
func CreateOrUpdateOAuthData(db database.Context, oauthData model.OAuthData) (model.OAuthData, error) {
	if oauthData.OAuthID == 0 {
		return model.OAuthData{}, ErrInvalidOAuthID
	}
	if len(oauthData.Data) == 0 {
		return model.OAuthData{}, ErrInvalidOAuthData
	}

	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result model.OAuthData
		err := gdb.
			Where(model.OAuthData{
				OAuthID: oauthData.OAuthID,
			}).
			Assign(oauthData).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.OAuthData{}, database.ErrInvalidDatabase
	}
}
