// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"reflect"
	"testing"
	"time"

	"github.com/condensat/bank-core/database/model"
)

func TestAddSwapInfo(t *testing.T) {
	const databaseName = "TestAddSwapInfo"
	t.Parallel()

	db := setup(databaseName, SwapModel())
	defer teardown(db, databaseName)

	swapRef, _ := AddSwap(db, model.SwapTypeAsk, 42, 101, 1.1, 102, 1.2)

	type args struct {
		swapID  model.SwapID
		status  model.SwapStatus
		payload model.Payload
	}
	tests := []struct {
		name    string
		args    args
		want    model.SwapInfo
		wantErr bool
	}{
		{"default", args{}, model.SwapInfo{}, true},
		{"valid", args{swapRef.ID, model.SwapStatusAccepted, "payload"}, createSwapInfo(swapRef.ID, model.SwapStatusAccepted, "payload"), false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddSwapInfo(db, tt.args.swapID, tt.args.status, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSwapInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("AddSwapInfo() wrong Timestamp %v", got.Timestamp)
				}
			}

			if !tt.wantErr {
				tt.want.ID = got.ID
				tt.want.Timestamp = got.Timestamp
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddSwapInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSwapInfo(t *testing.T) {
	const databaseName = "TestGetSwapInfo"
	t.Parallel()

	db := setup(databaseName, SwapModel())
	defer teardown(db, databaseName)

	swapRef, _ := AddSwap(db, model.SwapTypeAsk, 42, 101, 1.1, 102, 1.2)
	swapInfoRef, _ := AddSwapInfo(db, swapRef.ID, model.SwapStatusAccepted, "payload")
	if swapInfoRef.SwapID != swapRef.ID {
		t.Errorf("AddSwapInfo() wrnong swapID = %v, wantErr %v", swapInfoRef.SwapID, swapRef.ID)
		return
	}

	type args struct {
		swapInfoID model.SwapInfoID
	}
	tests := []struct {
		name    string
		args    args
		want    model.SwapInfo
		wantErr bool
	}{
		{"default", args{}, model.SwapInfo{}, true},
		{"valid", args{swapInfoRef.ID}, swapInfoRef, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSwapInfo(db, tt.args.swapInfoID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSwapInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSwapInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSwapInfoBySwapID(t *testing.T) {
	const databaseName = "TestGetSwapInfoBySwapID"
	t.Parallel()

	db := setup(databaseName, SwapModel())
	defer teardown(db, databaseName)

	swapRef, _ := AddSwap(db, model.SwapTypeAsk, 42, 101, 1.1, 102, 1.2)
	swapInfoRef, _ := AddSwapInfo(db, swapRef.ID, model.SwapStatusAccepted, "payload")
	if swapInfoRef.SwapID != swapRef.ID {
		t.Errorf("AddSwapInfo() wrnong swapID = %v, wantErr %v", swapInfoRef.SwapID, swapRef.ID)
		return
	}

	type args struct {
		swapID model.SwapID
	}
	tests := []struct {
		name    string
		args    args
		want    model.SwapInfo
		wantErr bool
	}{
		{"default", args{}, model.SwapInfo{}, true},
		{"valid", args{swapRef.ID}, swapInfoRef, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSwapInfoBySwapID(db, tt.args.swapID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSwapInfoBySwapID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSwapInfoBySwapID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createSwapInfo(swapID model.SwapID, status model.SwapStatus, payload model.Payload) model.SwapInfo {
	return model.SwapInfo{
		SwapID:  swapID,
		Status:  status,
		Payload: payload,
	}
}
