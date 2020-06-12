// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
	"time"
)

func TestBatchInfo_CryptoData(t *testing.T) {
	t.Parallel()

	type fields struct {
		ID        BatchInfoID
		Timestamp time.Time
		BatchID   BatchID
		Status    BatchStatus
		Type      DataType
		Data      BatchInfoData
	}
	tests := []struct {
		name    string
		fields  fields
		want    BatchInfoCryptoData
		wantErr bool
	}{
		{"default", fields{}, BatchInfoCryptoData{}, true},
		{"type", fields{Type: BatchInfoCrypto}, BatchInfoCryptoData{}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			p := &BatchInfo{
				ID:        tt.fields.ID,
				Timestamp: tt.fields.Timestamp,
				BatchID:   tt.fields.BatchID,
				Status:    tt.fields.Status,
				Type:      tt.fields.Type,
				Data:      tt.fields.Data,
			}
			got, err := p.CryptoData()
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchInfo.CryptoData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchInfo.CryptoData() = %v, want %v", got, tt.want)
			}
		})
	}
}
