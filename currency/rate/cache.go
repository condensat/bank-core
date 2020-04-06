// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/database/model"
)

type Rate struct {
	Timestamp time.Time
	Name      string
	Base      string
	Rate      float64
}

func (p *Rate) Encode() ([]byte, error) {
	return bank.EncodeObject(p)
}

func (p *Rate) Decode(data []byte) error {
	return bank.DecodeObject(data, bank.BankObject(p))
}

func UpdateRedisRate(ctx context.Context, currencyRates []model.CurrencyRate) {
	log := logger.Logger(ctx).WithField("Method", "rate.UpdateRedisRate")
	rdb := cache.ToRedis(appcontext.Cache(ctx))
	if rdb == nil {
		log.Error("Invalid redis cache")
		return
	}

	for _, r := range currencyRates {
		value := Rate{
			Timestamp: r.Timestamp,
			Name:      string(r.Name),
			Base:      string(r.Base),
			Rate:      float64(r.Rate),
		}

		key := formatRateKey(value.Name, value.Base)
		data, err := value.Encode()
		if err != nil {
			log.WithError(err).
				Error("Failed to encode object")
			continue
		}

		err = rdb.Set(key, data, 0).Err()
		if err != nil {
			log.WithError(err).
				Error("Failed to store rate to redis")
			continue
		}
	}
	log.
		WithField("Count", len(currencyRates)).
		Debug("Currency rate stored in redis cache")
}

func FetchRedisRate(ctx context.Context, name, base string) (Rate, error) {
	log := logger.Logger(ctx).WithField("Method", "rate.FetchRedisRate")
	rdb := cache.ToRedis(appcontext.Cache(ctx))
	if rdb == nil {
		log.Error("Invalid redis cache")
		return Rate{}, errors.New("Internal Error")
	}

	alias := name
	// 1 LBTC == 1 BTC
	if name == "LBTC" {
		alias = "BTC"
	}

	key := formatRateKey(alias, base)

	data, err := rdb.Get(key).Bytes()
	if err != nil {
		log.WithError(err).
			Error("Failed to fetch rate from redis")
		return Rate{}, errors.New("Internal Error")
	}

	var result Rate
	err = result.Decode(data)
	if err != nil {
		log.WithError(err).
			Error("Failed to decode object")
		return Rate{}, errors.New("Internal Error")
	}
	// override aliases
	result.Name = name
	return result, nil
}

func formatRateKey(name, base string) string {
	return fmt.Sprintf("rate:%s/%s", name, base)
}
