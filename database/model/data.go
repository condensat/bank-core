// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"encoding/json"
)

type DataInterface interface{}
type DataType String
type Data String

// EncodeData return Data from DataInterface struct. Encoded with json
func EncodeData(instance DataInterface) (Data, error) {
	if instance == nil {
		return Data(""), nil
	}
	data, err := json.Marshal(instance)
	if err != nil {
		return "", err
	}

	return Data(data), nil
}

// DecodeData store DataInterface from DataInterface struct. Decoded with json
func DecodeData(instance DataInterface, data Data) error {
	if len(data) == 0 {
		// NOOP
		return nil
	}
	return json.Unmarshal([]byte(data), instance)
}
