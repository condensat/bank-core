// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rate

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"encoding/json"

	"net/http"
	"net/url"

	"github.com/condensat/bank-core/database/model"
)

const (
	SourceName           = "oxr"
	OpenExchangeRatesURL = "https://openexchangerates.org"
	LatestPath           = "latest.json"
)

func FetchLatestRates(ctx context.Context, appID string) ([]model.CurrencyRate, error) {
	entryPoint := fmt.Sprintf("%s/api/%s", OpenExchangeRatesURL, LatestPath)
	u, err := url.Parse(entryPoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("app_id", appID)
	q.Add("prettyprint", "0")
	q.Add("show_alternative", "1")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "condensat/0.1")

	var httpClient http.Client
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseRate(string(body))
}

func parseRate(jsonBody string) ([]model.CurrencyRate, error) {
	var result []model.CurrencyRate

	var info struct {
		Disclaimer string                             `json:"disclaimer"`
		Licence    string                             `json:"licence"`
		Timestamp  int64                              `json:"timestamp"`
		Base       model.CurrencyName                 `json:"base"`
		Rates      map[model.CurrencyName]interface{} `json:"rates"`
	}

	err := json.Unmarshal([]byte(jsonBody), &info)
	if err != nil {
		return nil, err
	}

	for name, value := range info.Rates {
		switch rate := value.(type) {
		case float64:
			result = append(result, model.CurrencyRate{
				Timestamp: time.Unix(info.Timestamp, 0).UTC(),
				Source:    SourceName,
				Base:      info.Base,
				Name:      name,
				Rate:      model.CurrencyRateValue(rate),
			})

		default:
			continue
		}
	}

	return result, nil
}
