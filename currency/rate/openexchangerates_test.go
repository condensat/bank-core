// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package rate

import (
	"context"
	"reflect"
	"testing"

	"github.com/condensat/bank-core/database/model"
)

func TestFetchLatestRates(t *testing.T) {
	// ctx := context.TODO()
	type args struct {
		ctx   context.Context
		appID string
	}
	tests := []struct {
		name    string
		args    args
		want    []model.Currency
		wantErr bool
	}{
		// modify app_id in args
		// {"Fetch", args{ctx, "app_id"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchLatestRates(tt.args.ctx, tt.args.appID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchLatestRates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchLatestRates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseRate(t *testing.T) {
	type args struct {
		jsonBody string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"parse", args{mockJsonBody}, 193, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRate(tt.args.jsonBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("parseRate() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

const (
	mockJsonBody = `
{
  "disclaimer": "Usage subject to terms: https://openexchangerates.org/terms",
  "license": "https://openexchangerates.org/license",
  "timestamp": 1584547200,
  "base": "USD",
  "rates": {
    "AED": 3.6732,
    "AFN": 75.949999,
    "ALL": 111.525,
    "AMD": 490.158769,
    "ANG": 1.789649,
    "AOA": 506.3135,
    "ARS": 63.3151,
    "AUD": 1.718006,
    "AWG": 1.8,
    "AZN": 1.7025,
    "BAM": 1.781371,
    "BBD": 2,
    "BDT": 84.815324,
    "BGN": 1.800904,
    "BHD": 0.377432,
    "BIF": 1896,
    "BMD": 1,
    "BND": 1.438727,
    "BOB": 6.883712,
    "BRL": 5.116913,
    "BSD": 1,
    "BTC": 0.000186671797,
    "BTN": 74.218027,
    "BTS": 68.8021162362,
    "BWP": 11.673437,
    "BYN": 2.390967,
    "BZD": 2.015366,
    "CAD": 1.453075,
    "CDF": 1706,
    "CHF": 0.972875,
    "CLF": 0.031276,
    "CLP": 865.599013,
    "CNH": 7.079949,
    "CNY": 7.0475,
    "COP": 4137.88,
    "CRC": 566.18116,
    "CUC": 1,
    "CUP": 25.75,
    "CVE": 100.55,
    "CZK": 25.549389,
    "DASH": 0.0199855656,
    "DJF": 178.025,
    "DKK": 6.899248,
    "DOGE": 621.51245,
    "DOP": 53.955,
    "DZD": 121.63,
    "EAC": 2867.98535556,
    "EGP": 15.7473,
    "EMC": 1.2125648612,
    "ERN": 14.999767,
    "ETB": 32.782814,
    "ETH": 0.0085969739,
    "EUR": 0.923109,
    "FCT": 0.5715065711,
    "FJD": 2.28,
    "FKP": 0.850125,
    "FTC": 22.4451027826,
    "GBP": 0.850125,
    "GEL": 3.075,
    "GGP": 0.850125,
    "GHS": 5.59,
    "GIP": 0.850125,
    "GMD": 50.9,
    "GNF": 9412.5,
    "GTQ": 7.647643,
    "GYD": 208.507695,
    "HKD": 7.768683,
    "HNL": 24.91,
    "HRK": 7.010765,
    "HTG": 94.790722,
    "HUF": 326.546876,
    "IDR": 15504.5,
    "ILS": 3.762215,
    "IMP": 0.850125,
    "INR": 74.981251,
    "IQD": 1190,
    "IRR": 42105,
    "ISK": 140.299964,
    "JEP": 0.850125,
    "JMD": 135.290365,
    "JOD": 0.7095,
    "JPY": 108.436,
    "KES": 104.21,
    "KGS": 71.35925,
    "KHR": 4045,
    "KMF": 448.574861,
    "KPW": 900,
    "KRW": 1260.7,
    "KWD": 0.31034,
    "KYD": 0.833212,
    "KZT": 443.524728,
    "LAK": 8894.738566,
    "LBP": 1528.29372,
    "LD": 320,
    "LKR": 185.06419,
    "LRD": 198.049985,
    "LSL": 16.59,
    "LTC": 0.0294464075,
    "LYD": 1.405,
    "MAD": 9.64725,
    "MDL": 17.849263,
    "MGA": 3700,
    "MKD": 56.010837,
    "MMK": 1428.871499,
    "MNT": 2757.604638,
    "MOP": 7.997322,
    "MRO": 357,
    "MRU": 37.7,
    "MUR": 39.300706,
    "MVR": 15.41,
    "MWK": 735,
    "MXN": 23.70862,
    "MYR": 4.3735,
    "MZN": 65.999999,
    "NAD": 16.59,
    "NGN": 366.5,
    "NIO": 34.2,
    "NMC": 3.2262324937,
    "NOK": 11.090293,
    "NPR": 118.749213,
    "NVC": 0.3852517642,
    "NXT": 129.481760417,
    "NZD": 1.731102,
    "OMR": 0.385159,
    "PAB": 1,
    "PEN": 3.544,
    "PGK": 3.4075,
    "PHP": 51.5705,
    "PKR": 158.45,
    "PLN": 4.179699,
    "PPC": 2.5433167196,
    "PYG": 6579.780947,
    "QAR": 3.641,
    "RON": 4.4815,
    "RSD": 108.455,
    "RUB": 80.1835,
    "RWF": 940,
    "SAR": 3.753676,
    "SBD": 8.267992,
    "SCR": 13.705328,
    "SDG": 55.275,
    "SEK": 10.24267,
    "SGD": 1.445165,
    "SHP": 0.850125,
    "SLL": 7602.997835,
    "SOS": 585,
    "SRD": 7.458,
    "SSP": 130.26,
    "STD": 22134.769315,
    "STN": 22.3,
    "STR": 27.7460915179,
    "SVC": 8.749072,
    "SYP": 515.17941,
    "SZL": 16.59,
    "THB": 32.455826,
    "TJS": 9.740177,
    "TMT": 3.51,
    "TND": 2.8825,
    "TOP": 2.3723,
    "TRY": 6.480193,
    "TTD": 6.755347,
    "TWD": 30.387499,
    "TZS": 2304.6,
    "UAH": 27.307171,
    "UGX": 3749.276568,
    "USD": 1,
    "UYU": 43.480319,
    "UZS": 9512.5,
    "VEF": 248487.642241,
    "VEF_BLKMKT": 75724.82,
    "VEF_DICOM": 74007.97,
    "VEF_DIPRO": 4323282,
    "VES": 73882.308762,
    "VND": 23588.486528,
    "VTC": 6.8084886676,
    "VUV": 118.953262,
    "WST": 2.715397,
    "XAF": 605.519582,
    "XAG": 0.08257689,
    "XAU": 0.00066984,
    "XCD": 2.70255,
    "XDR": 0.729358,
    "XMR": 0.0279122776,
    "XOF": 605.519582,
    "XPD": 0.00061307,
    "XPF": 110.156164,
    "XPM": 21.0235762811,
    "XPT": 0.00160773,
    "XRP": 6.8473644877,
    "YER": 250.349961,
    "ZAR": 17.108099,
    "ZMW": 16.571972,
    "ZWL": 322.000001
  }
}`
)
