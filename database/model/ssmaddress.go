// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type SsmAddressID ID
type SsmPublicAddress String
type SsmPubkey String
type SsmBlindingKey String

type SsmAddress struct {
	ID            SsmAddressID     `gorm:"primary_key;"`                   // [PK] SsmAddress ID
	PublicAddress SsmPublicAddress `gorm:"unique_index;not null;size:126"` // Ssm Address, non mutable
	ScriptPubkey  SsmPubkey        `gorm:"not null;size:66"`               // Ssm Script, non mutable
	BlindingKey   SsmBlindingKey   `gorm:"not null;size:64"`               // Ssm BlindingKey, non mutable (optional)
}

func (p *SsmAddress) IsValid() bool {
	return len(p.PublicAddress) > 0 && len(p.ScriptPubkey) > 0
}
