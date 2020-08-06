// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/ssm/commands"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/sirupsen/logrus"
)

const (
	SsmMaxUnusedAddress = 1000
	SsmFillBatchSize    = 10
)

type SsmInfo struct {
	Device      string
	Chain       string
	Fingerprint string
}

// SsmPool
func SsmPool(ctx context.Context, epoch time.Time, infos []SsmInfo) {
	log := logger.Logger(ctx).WithField("Method", "task.SsmPool")
	db := appcontext.Database(ctx)

	log.WithFields(logrus.Fields{
		"Epoch": epoch.Truncate(time.Millisecond),
		"Count": len(infos),
	}).Info("Maintain ssm pool addresses")

	for _, info := range infos {
		ssm := common.SsmClientFromContext(ctx, info.Device)

		unusedCount, err := database.CountSsmAddressByState(db,
			model.SsmChain(info.Chain),
			model.SsmFingerprint(info.Fingerprint),
			model.SsmAddressStatusUnused,
		)
		if err != nil {
			log.WithError(err).Error("CountSsmAddressByState failed")
			continue
		}

		// count actual ssm addresses count for chain/fingerprint
		addressCount, err := database.CountSsmAddress(db,
			model.SsmChain(info.Chain),
			model.SsmFingerprint(info.Fingerprint),
		)
		if err != nil {
			log.WithError(err).Error("CountSsmAddress failed")
			continue
		}

		log.WithFields(logrus.Fields{
			"UnusedCount":  unusedCount,
			"AddressCount": addressCount,
		}).Debug("SsmPool status")

		// Fill ssm pool
		nextUnusedCount := unusedCount + SsmFillBatchSize
		if nextUnusedCount > SsmMaxUnusedAddress {
			nextUnusedCount = SsmMaxUnusedAddress
		}
		for unusedCount < nextUnusedCount {
			err := func() error {
				// create new address for next path
				// Todo: manage annual rotation for path
				nextPath := fmt.Sprintf("84h/0h/%d", addressCount+1)

				ssmAddress, err := ssm.NewAddress(ctx, commands.SsmPath{
					Chain:       info.Chain,
					Fingerprint: info.Fingerprint,
					Path:        nextPath,
				})
				if err != nil {
					log.WithError(err).Error("NewAddress failed")
					return err
				}
				if info.Chain != ssmAddress.Address {
					if err != nil {
						log.WithError(err).Error("Wrong ssmAddress chain")
						return err
					}
				}

				// store new address to database
				ssmAddressID, err := database.AddSsmAddress(db,
					model.SsmAddress{
						PublicAddress: model.SsmPublicAddress(ssmAddress.Address),
						ScriptPubkey:  model.SsmPubkey(ssmAddress.PubKey),
						BlindingKey:   model.SsmBlindingKey(ssmAddress.BlindingKey),
					},
					model.SsmAddressInfo{
						Chain:       model.SsmChain(info.Chain),
						Fingerprint: model.SsmFingerprint(info.Fingerprint),
						HDPath:      model.SsmHDPath(nextPath),
					},
				)
				if err != nil {
					log.WithError(err).Error("AddSsmAddress failed")
					return err
				}

				log.
					WithField("ssmAddressID", ssmAddressID).
					Debug("New ssm address")

				return nil
			}()
			if err != nil {
				break
			}

			unusedCount++
			addressCount++
		}
	}
}
