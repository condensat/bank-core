// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"
)

const (
	AssetInfoEndpoint = "https://assets.blockstream.info/"
	AssetIconEndpoint = "https://assets.blockstream.info/icons.json"
)

// UpdateAssetInfo
func UpdateAssetInfo(ctx context.Context, epoch time.Time) {
	processAssetInfo(ctx)
	processAssetIcon(ctx)
}

func processAssetInfo(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "tasks.processAssetInfo")
	db := appcontext.Database(ctx)

	jsonData, err := fetchEnpoint(ctx, AssetInfoEndpoint)
	if err != nil {
		log.WithError(err).
			Error("Failed to fetch AssetInfo")
		return
	}

	assetInfos, err := parseAssetInfo(jsonData)
	if err != nil {
		log.WithError(err).
			Error("Failed to parseAssetInfo")
		return
	}

	for _, assetInfo := range assetInfos {
		assetHash := model.AssetHash(assetInfo.AssetID)

		// check if Asset exists
		if database.AssetHashExists(db, assetHash) {
			continue
		}

		// create Asset - use Ticker as CurrencyName
		asset, err := database.AddAsset(db, assetHash, model.CurrencyName(assetInfo.Ticker))
		if err != nil {
			log.WithError(err).
				Error("Failed to AddOrUpdateAssetInfo")
			continue
		}

		// add AssetInfo
		_, err = database.AddOrUpdateAssetInfo(db, model.AssetInfo{
			AssetID:   asset.ID,
			Domain:    assetInfo.Entity.Domain,
			Name:      assetInfo.Name,
			Ticker:    assetInfo.Ticker,
			Precision: uint8(assetInfo.Precision),
		})
		if err != nil {
			log.WithError(err).
				WithField("AssetInfo", model.AssetInfo{
					AssetID:   asset.ID,
					Domain:    assetInfo.Entity.Domain,
					Name:      assetInfo.Name,
					Ticker:    assetInfo.Ticker,
					Precision: uint8(assetInfo.Precision),
				}).
				Error("Failed to AddOrUpdateAssetInfo")
			continue
		}
	}
}

func processAssetIcon(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "tasks.processAssetIcon")
	db := appcontext.Database(ctx)

	jsonData, err := fetchEnpoint(ctx, AssetIconEndpoint)
	if err != nil {
		log.WithError(err).
			Error("Failed to fetchEnpoint")
		return
	}

	assetIcons, err := parseAssetIcon(jsonData)
	if err != nil {
		log.WithError(err).
			Error("Failed to parseAssetIcon")
		return
	}

	for _, assetIcon := range assetIcons {
		assetHash := model.AssetHash(assetIcon.AssetID)

		if !database.AssetHashExists(db, assetHash) {
			if assetHash != PolicyAssetLiquid {
				log.Error("Asset not found")
				continue
			}

			_, err = database.AddAsset(db, assetHash, model.CurrencyName(TickerAssetLiquid))
			if err != nil {
				log.WithError(err).
					Error("Failed to AddAsset")
				continue
			}
		}

		asset, err := database.GetAssetByHash(db, assetHash)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetAssetByHash")
			continue
		}

		_, err = database.AddOrUpdateAssetIcon(db, model.AssetIcon{
			AssetID: asset.ID,
			Data:    assetIcon.Data,
		})
		if err != nil {
			log.WithError(err).
				Error("Failed to AddOrUpdateAssetIcon")
			continue
		}
	}
}

func parseAssetInfo(jsonData []byte) ([]AssetInfo, error) {
	infos := make(map[string]AssetInfo)
	err := json.Unmarshal(jsonData, &infos)
	if err != nil {
		return nil, err
	}

	result := make([]AssetInfo, 0)
	for _, info := range infos {
		result = append(result, info)
	}
	return result, nil
}

func parseAssetIcon(jsonData []byte) ([]AssetIcon, error) {
	icons := make(map[string]string)
	err := json.Unmarshal(jsonData, &icons)
	if err != nil {
		return nil, err
	}

	result := make([]AssetIcon, 0)
	for assetID, icon := range icons {
		data, err := base64.StdEncoding.DecodeString(icon)
		if err != nil {
			continue
		}
		result = append(result, AssetIcon{
			AssetID: assetID,
			Data:    data,
		})
	}
	return result, nil
}

func fetchEnpoint(ctx context.Context, enpoint string) ([]byte, error) {
	resp, err := http.Get(enpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type AssetIcon struct {
	AssetID string `json:"asset_id"`
	Data    []byte
}

type AssetInfo struct {
	AssetID  string `json:"asset_id"`
	Contract struct {
		Entity struct {
			Domain string `json:"domain"`
		} `json:"entity"`
		IssuerPubkey string `json:"issuer_pubkey"`
		Name         string `json:"name"`
		Nonce        string `json:"nonce"`
		Precision    int    `json:"precision"`
		Ticker       string `json:"ticker"`
		Version      int    `json:"version"`
	} `json:"contract"`
	IssuanceTxin struct {
		Txid string `json:"txid"`
		Vin  int    `json:"vin"`
	} `json:"issuance_txin"`
	IssuancePrevout struct {
		Txid string `json:"txid"`
		Vout int    `json:"vout"`
	} `json:"issuance_prevout"`
	Name      string `json:"name"`
	Ticker    string `json:"ticker"`
	Precision int    `json:"precision"`
	Entity    struct {
		Domain string `json:"domain"`
	} `json:"entity"`
	Version      int    `json:"version"`
	IssuerPubkey string `json:"issuer_pubkey"`
}
