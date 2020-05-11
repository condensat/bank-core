// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"io"
	"sync"

	"github.com/condensat/bank-core/api/services/assets"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/nfnt/resize"
)

const (
	IconSize = uint(128)
)

var icons map[string][]byte
var iconsMutex sync.Mutex

func init() {
	icons = make(map[string][]byte)
}

func getTickerIcon(ctx context.Context, ticker string) []byte {
	iconsMutex.Lock()
	defer iconsMutex.Unlock()

	// lookup in cache
	if icon, ok := icons[ticker]; ok {
		return icon
	}

	// create icon & add to cache
	icon := getTickerIconNoCache(ctx, ticker)
	if icon != nil {
		icons[ticker] = icon
	}
	return icon
}

func getTickerIconNoCache(ctx context.Context, ticker string) []byte {
	db := appcontext.Database(ctx)

	switch ticker {
	case "CHF":
		return resizeIcon(IconSize, dataFromBase64(assets.CHFIcon))

	case "EUR":
		return resizeIcon(IconSize, dataFromBase64(assets.EuroIcon))

	case "BTC":
		return resizeIcon(IconSize, dataFromBase64(assets.BitcoinIcon))

	case "TBTC":
		return resizeIcon(IconSize, dataFromBase64(assets.BitcoinTestnetIcon))

	case "LBTC":
		// LBTC is not listed from blockstream API but have an icon
		// must be assetID = 1 in database
		const liquidAssetID = model.AssetID(1)
		if assetIcon, err := database.GetAssetIcon(db, liquidAssetID); err == nil {
			return resizeIcon(IconSize, assetIcon.Data)
		}
		return nil

	default:

		if asset, err := database.GetAssetByCurrencyName(db, model.CurrencyName(ticker)); err == nil {
			if assetIcon, err := database.GetAssetIcon(db, asset.ID); err == nil {
				return resizeIcon(IconSize, assetIcon.Data)
			}
		}
		return nil
	}
}

func dataFromBase64(b64 string) []byte {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil
	}
	return data
}

func resizeIcon(width uint, data []byte) []byte {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}
	m := resize.Resize(width, 0, img, resize.Bilinear)

	var b bytes.Buffer
	out := io.Writer(&b)
	err = png.Encode(out, m)
	if err != nil {
		return data
	}

	return b.Bytes()
}
